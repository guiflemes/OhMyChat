package telegram

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"notion-agenda/settings"
)

type Action interface {
	//checkar se precisa add nodes ah mais por exemplo:
	//add faturas por add para virar opções dentro do node
	Execute(message string) string
}

type message struct {
	id      string
	parent  string
	content string
	action  Action
}

func (m *message) HasAction() bool {
	return m.action != nil
}

type messageNode struct {
	firstChild  *messageNode
	nextSibling *messageNode
	message     message
}

//                           coco
//     xixi (coco)                     veve(coco)     lolo(coco)
// dudu(xixi)   caca(xixi)                               didi(lolo)

func (n *messageNode) insert(node *messageNode) {
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

	if n.firstChild.message.id == node.message.parent {
		n.firstChild.insert(node)
		return
	}

	found := false
	sibling := n.firstChild
	for sibling.nextSibling != nil {
		sibling = sibling.nextSibling
		if sibling.message.id == node.message.parent {
			found = !found
			break
		}
	}

	if !found {
		fmt.Printf(
			"node %s without parent, given parent %s not found/n",
			node.message.id,
			node.message.parent,
		)
		return
	}

	sibling.insert(node)

}

func (n *messageNode) searchOneLevel(id string) *messageNode {
	if n.message.id == id {
		return n
	}
	return n.searchChild(id)
}

func (n *messageNode) searchChild(id string) *messageNode {
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

func (n *messageNode) transverseInChildren(fn func(child *messageNode)) {
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

func (n *messageNode) repChildren() string {
	rep := ""
	count := 1
	n.transverseInChildren(func(child *messageNode) {
		rep += fmt.Sprintf("%d: %s\n", count, child.message.id)
		count++
	})
	return rep
}

type MessageTree struct {
	root *messageNode
}

func (t *MessageTree) Insert(node *messageNode) *MessageTree {
	if t.root == nil {
		t.root = node
		return t
	}

	t.root.insert(node)
	return t
}

func (t *MessageTree) Search(id string) *messageNode {
	if t.root == nil {
		return nil
	}

	return t.root.searchChild(id)
}

type myAction struct{}

func (m *myAction) Execute(message string) string {
	invoices := ""
	for i := 0; i <= 5; i++ {
		invoice := fmt.Sprintf("www.minhasfaturas.com/vencidas/12%d", i)
		invoices += invoice + "\n"
	}
	return invoices
}

func Fn() *MessageTree {

	tree := &MessageTree{}
	tree.Insert(&messageNode{message: message{parent: "", id: "coco", content: "O que voce gostaria de saber?"}}).
		Insert(&messageNode{message: message{parent: "coco", id: "faturas", content: "Fatura, escolha as opções"}}).
		Insert(&messageNode{message: message{parent: "coco", id: "assinaturas", content: "Assinaturas, esolhas as opções"}}).
		Insert(&messageNode{message: message{parent: "coco", id: "marvin", content: "Marvin, escolha o role"}}).
		Insert(&messageNode{message: message{parent: "faturas", id: "atrasadas", content: "ok, verificando", action: &myAction{}}}).
		Insert(&messageNode{message: message{parent: "faturas", id: "pagas"}}).
		Insert(&messageNode{message: message{parent: "marvin", id: "coco"}})

	//tree.root.firstChild.printChildren()

	//tree.root.firstChild.nextSibling.nextSibling.printChildren()
	//fmt.Println(tree.root.firstChild.nextSibling.nextSibling.repChildren())
	//fmt.Println(tree.root.firstChild.searchChild("caca"))
	return tree
}

type commandEngine struct {
	tree         *MessageTree
	node         *messageNode
	dialogLaunch bool
	actionQueue  ActionQueue
}

func (e *commandEngine) IsInitialized() bool {
	return e.tree != nil || e.node != nil
}

func (e *commandEngine) resolveMessageNode(messageID string) {

	if e.dialogLaunch {
		node := e.node.searchOneLevel(messageID)

		if node == nil {
			e.node = e.tree.root
			e.dialogLaunch = false
			return
		}
		e.node = node
		return
	}

	e.dialogLaunch = true

}

func (e *commandEngine) Reply(chatID int64, messageID string) tgbotapi.MessageConfig {

	e.resolveMessageNode(messageID)

	if e.node.message.HasAction() {
		go func() {
			content := e.node.message.action.Execute(messageID)
			e.actionQueue <- tgbotapi.NewMessage(chatID, content)
		}()
	}

	buttons := make([]tgbotapi.KeyboardButton, 0)

	e.node.transverseInChildren(func(child *messageNode) {
		buttons = append(buttons, tgbotapi.NewKeyboardButton(child.message.id))
	})

	keyboard := tgbotapi.NewReplyKeyboard(buttons)

	msg := tgbotapi.NewMessage(chatID, e.node.message.content)
	msg.ReplyMarkup = keyboard
	return msg

}

type ActionQueue chan tgbotapi.MessageConfig

type WorkFlowEngine struct {
	client           *tgbotapi.BotAPI
	notRecognizedMsg string
	dialogLaunch     bool
	unmarshalMsg     func(msg string) message
	commandEngine    *commandEngine
	actionQueue      ActionQueue
}

func NewEngine() *WorkFlowEngine {
	client, err := tgbotapi.NewBotAPI(settings.GETENV("TELEGRAM_TOKEN"))

	if err != nil {
		log.Panic(err)
	}

	commandTree := Fn()
	actionQueue := make(chan tgbotapi.MessageConfig)

	return &WorkFlowEngine{
		client:           client,
		notRecognizedMsg: "I did not understand, can you repeat?",
		dialogLaunch:     false,
		commandEngine: &commandEngine{
			tree:        commandTree,
			node:        commandTree.root,
			actionQueue: actionQueue,
		},
		actionQueue: actionQueue,
	}
}

func (e *WorkFlowEngine) HasPostback() bool {
	return e.commandEngine.IsInitialized()
}

func (e *WorkFlowEngine) Chating(timeout int) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = timeout

	updates := e.client.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		replyMsg := e.commandEngine.Reply(update.Message.Chat.ID, update.Message.Text)

		log.Printf("message %s", update.Message.Text)

		replyMsg.ReplyToMessageID = update.Message.MessageID

		_, err := e.client.Send(replyMsg)
		if err != nil {
			log.Println(err)
		}

	}

}

func (e *WorkFlowEngine) ChatingWithAction(timeout int) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = timeout

	updates := e.client.GetUpdatesChan(u)

	for {
		select {
		case update := <-updates:
			e.replyMessage(update)

		case action := <-e.actionQueue:
			_, err := e.client.Send(action)
			if err != nil {
				log.Println(err)
			}
		}
	}

}

func (e *WorkFlowEngine) replyMessage(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	replyMsg := e.commandEngine.Reply(update.Message.Chat.ID, update.Message.Text)
	log.Printf("message %s", update.Message.Text)
	replyMsg.ReplyToMessageID = update.Message.MessageID

	_, err := e.client.Send(replyMsg)
	if err != nil {
		log.Println(err)
	}

}
