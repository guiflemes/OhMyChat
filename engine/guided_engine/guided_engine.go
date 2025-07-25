package guidedengine

import (
	"fmt"

	"golang.org/x/net/context"

	"github.com/guiflemes/ohmychat/message"
)

type Action interface {
	Handle(ctx context.Context, message *message.Message) error
}

type ActionReplyPair struct {
	ReplyTo chan<- message.Message
	Action  Action
	Input   message.Message
}

type ActionStorageService interface {
	Enqueue(actioonPair ActionReplyPair)
}

type Message struct {
	id      string
	parent  string
	name    string
	Content string
	Action  Action
}

func (m Message) ID() string {
	return m.id
}

func (m Message) HasAction() bool {
	return m.Action != nil
}

type MessageNode struct {
	firstChild  *MessageNode
	nextSibling *MessageNode
	message     Message
}

func NewMessageNode(
	id string,
	parent string,
	name string,
	content string,
	action Action,
) *MessageNode {
	return &MessageNode{message: Message{
		parent:  parent,
		id:      id,
		name:    name,
		Content: content,
		Action:  action,
	}}
}

func (n *MessageNode) Message() Message {
	return n.message
}

func (n *MessageNode) insert(node *MessageNode) {

	if n.message.id == node.message.parent {
		if n.firstChild == nil {
			n.firstChild = node
			return
		}
		sibling := n.firstChild
		for sibling.nextSibling != nil {
			sibling = sibling.nextSibling
		}
		sibling.nextSibling = node
		return
	}

	if n.firstChild != nil {
		if n.firstChild.message.id == node.message.parent {
			n.firstChild.insert(node)
			return
		}
		sibling := n.firstChild
		for sibling != nil {
			if sibling.message.id == node.message.parent {
				sibling.insert(node)
				return
			}
			if sibling.firstChild != nil {
				sibling.firstChild.insert(node)
			}
			sibling = sibling.nextSibling
		}
	}

}

func (n *MessageNode) SearchOneLevel(id string) *MessageNode {
	if n.message.id == id {
		return n
	}
	return n.searchChild(id)
}

func (n *MessageNode) searchChild(id string) *MessageNode {
	if n.firstChild == nil {
		return nil
	}

	child := n.firstChild

	if child.message.id == id {
		return child
	}

	for child.nextSibling != nil {
		child = child.nextSibling
		if child.message.id == id {
			return child
		}
	}

	return nil
}

func (n *MessageNode) TransverseInChildren(fn func(child *MessageNode)) {
	if n.firstChild == nil {
		return
	}

	child := n.firstChild
	if child.nextSibling != nil {
		for child != nil {
			fn(child)
			child = child.nextSibling
		}
		return
	}

	fn(child)

}

func (n *MessageNode) RepChildren() string {
	rep := ""
	count := 1
	n.TransverseInChildren(func(child *MessageNode) {
		rep += fmt.Sprintf("%d: %s\n", count, child.message.id)
		count++
	})
	return rep
}

type MessageTree struct {
	root *MessageNode
}

func (t *MessageTree) Root() *MessageNode {
	return t.root
}

func (t *MessageTree) SetRoot(root *MessageNode) {
	t.root = root
}

func (t *MessageTree) Insert(node *MessageNode) *MessageTree {
	if t.root == nil {
		t.root = node
		return t
	}

	t.root.insert(node)
	return t
}

func (t *MessageTree) Search(id string) *MessageNode {
	if t.root == nil {
		return nil
	}

	return t.root.searchChild(id)
}

type ChatRoutingRule int

const (
	Fallback ChatRoutingRule = iota
	KeepContext
	HumanHandOff
)

type GuidedResponseRepo interface {
	GetMessageTree(workflowID string) (*MessageTree, error)
}

// guidedResponseEngine is currently not operational and pending implementation.
type guidedResponseEngine struct {
	tree         *MessageTree
	node         *MessageNode
	dialogLaunch bool
	actionQueue  ActionStorageService
	setup        bool
	chatRouting  ChatRoutingRule
	repo         GuidedResponseRepo
}

// Deprecated: guidedResponseEngine is currently not operational and pending implementation.
func NewGuidedResponseEngine(
	actionQueue ActionStorageService,
	repo GuidedResponseRepo,
) *guidedResponseEngine {

	return &guidedResponseEngine{
		actionQueue: actionQueue,
		chatRouting: Fallback,
		repo:        repo,
	}
}

func (e *guidedResponseEngine) IsReady() bool {
	return e.setup
}

func (e *guidedResponseEngine) Config(workflowID string) error {
	tree, err := e.repo.GetMessageTree(workflowID)
	if err != nil {
		return err
	}

	e.tree = tree
	e.node = tree.Root()
	e.setup = true
	return nil
}

func (e *guidedResponseEngine) route() {

	if e.chatRouting == Fallback {
		e.node = e.tree.Root()
	}

	if e.chatRouting == KeepContext || e.chatRouting == HumanHandOff {
		e.node = e.tree.Root()
	}

}

func (e *guidedResponseEngine) resolveMessageNode(messageID string) {

	if e.dialogLaunch {
		node := e.node.SearchOneLevel(messageID)

		if node == nil {
			e.route()
			return
		}
		e.node = node
		return
	}

	e.dialogLaunch = true

}

func (e *guidedResponseEngine) GetActionStorageService() ActionStorageService {
	return e.actionQueue
}

func (e *guidedResponseEngine) Name() string {
	return "guided"
}

func (e *guidedResponseEngine) HandleMessage(ctx context.Context, input message.Message, output chan<- message.Message) {

	if !e.setup {
		response := &input
		response.Output = "some error ocurred, please contant admin"
		output <- *response
		return
	}

	e.resolveMessageNode(input.Input)

	if e.node.Message().HasAction() {
		actionPair := ActionReplyPair{ReplyTo: output, Action: e.node.message.Action, Input: input}
		storageAction := e.GetActionStorageService()
		storageAction.Enqueue(actionPair)
	}

	options := make([]message.Option, 0)
	e.node.TransverseInChildren(func(child *MessageNode) {
		options = append(
			options,
			message.Option{ID: child.Message().ID(), Name: child.Message().name},
		)
	})

	response := &input
	response.Output = e.node.Message().Content
	response.Options = options
	response.ResponseType = message.OptionResponse
	output <- *response
}
