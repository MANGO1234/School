package gui

import (
	"../buffer"
	. "../common"
	"../documentmanager"
	"../network"
	"../version"
	"fmt"
	"github.com/nsf/termbox-go"
	"github.com/satori/go.uuid"
	"sort"
	"strconv"
)

// States
const STATE_MENU = 0
const STATE_MENU_RETRY = 1
const STATE_EXIT = 10
const STATE_DOCUMENT = 20
const STATE_CONNECT = 30
const STATE_ERROR = 40

// Menu Options
const OPTION_EXIT = "Exit"
const OPTION_CONNECT = "Connect"
const OPTION_DISCONNECT = "Disconnect"
const OPTION_NEW_DOCUMENT = "New Document"
const OPTION_CLOSE_DOCUMENT = "Close Document"

var appState struct {
	State       int
	DocModel    *documentmanager.DocumentModel
	ScreenY     int
	MenuOptions []string
	Manager     *network.NetworkManager
	TempData    interface{}
}

func doAction(input string) {
	if appState.State == STATE_MENU || appState.State == STATE_MENU_RETRY {
		n, err := strconv.Atoi(input)
		if err != nil || n < 1 || n > len(appState.MenuOptions) {
			appState.State = STATE_MENU_RETRY
			return
		}
		if appState.MenuOptions[n-1] == OPTION_EXIT {
			appState.State = STATE_EXIT
		} else if appState.MenuOptions[n-1] == OPTION_CONNECT {
			appState.State = STATE_CONNECT
		} else if appState.MenuOptions[n-1] == OPTION_DISCONNECT {
			appState.Manager.Disconnect()
			if err != nil {
				appState.State = STATE_ERROR
				appState.TempData = err
			}
		} else if appState.MenuOptions[n-1] == OPTION_NEW_DOCUMENT {
			appState.DocModel = newDocument(StringToSiteId(uuid.NewV1().String()))
			appState.ScreenY = 0
			appState.State = STATE_DOCUMENT
		} else if appState.MenuOptions[n-1] == OPTION_CLOSE_DOCUMENT {
			appState.Manager.CompleteDisconnect()
			appState.DocModel = nil
		} else {
			appState.State = STATE_MENU_RETRY
		}
	} else if appState.State == STATE_CONNECT {
		err := appState.Manager.ConnectTo(input)
		if err != nil {
			appState.State = STATE_ERROR
			appState.TempData = err
		} else {
			appState.State = STATE_MENU
		}
	} else if appState.State == STATE_ERROR {
		appState.State = STATE_MENU
	}
}

func getPrompt() *buffer.Prompt {
	if appState.State == STATE_MENU || appState.State == STATE_MENU_RETRY {
		options := make([]string, 0, 10)
		if appState.DocModel == nil {
			options = append(options, OPTION_NEW_DOCUMENT)
		} else {
			options = append(options, OPTION_CONNECT)
			options = append(options, OPTION_DISCONNECT)
			options = append(options, OPTION_CLOSE_DOCUMENT)
		}
		options = append(options, OPTION_EXIT)
		appState.MenuOptions = options

		str := ""
		for i, option := range options {
			str += strconv.Itoa(i+1) + ". " + option + "\n"
		}
		str += "\n"
		str += "Esc to switch between menu and editing the document\n"
		str += "\nPeers In Network:\n"
		netMeta := appState.Manager.GetNetworkMetadata().ToList()
		sort.Sort(netMeta)
		for _, node := range netMeta {
			str += node.Id + ": Addr = " + node.Addr + ", Left = " + strconv.FormatBool(node.Left) + "\n"
		}
		str += "\n"
		if appState.State == STATE_MENU_RETRY {
			str += "Input is not a valid input, please try again: "
		} else {
			str += "Enter number to exectute option: "
		}
		return buffer.NewPrompt(str)
	} else if appState.State == STATE_CONNECT {
		return buffer.NewPrompt("Enter ip to connect to: ")
	} else if appState.State == STATE_ERROR {
		err := appState.TempData.(error)
		str := fmt.Sprintf("%v\n\nPress Enter to continue", err)
		return buffer.NewPrompt(str)
	}
	panic("UNKNOWN STATE " + strconv.Itoa(appState.State))
}

func StartMainLoop(manager *network.NetworkManager) {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	appState.State = STATE_MENU
	appState.Manager = manager
	appState.Manager.SetRemoteOpHandler(func(msg []byte) {
		if appState.DocModel != nil {
			ops := documentmanager.RemoteOperationsFromSlice(msg)
			for _, op := range ops {
				appState.DocModel.ApplyRemoteOperation(op)
			}
		}
	})
	appState.Manager.SetGetOpsReceiveVersion(func() []byte {
		if appState.DocModel != nil {
			return appState.DocModel.GetVersionVectorReceived().MarshalJSON()
		} else {
			return nil
		}
	})
	appState.Manager.SetVersionCheckHandler(func(data []byte) ([]byte, bool) {
		err, vector := version.UnmarshalJSON(data)
		if appState.DocModel != nil && err == nil {
			ops, queueOps := appState.DocModel.GetMissingOperations(vector)
			ops = append(ops, queueOps...)
			if len(ops) > 0 {
				return documentmanager.RemoteOperationsToSlice(ops), true
			}
		}
		return nil, false
	})
	for {
		if appState.State == STATE_EXIT {
			appState.Manager.Disconnect()
			break
		}
		if appState.State == STATE_DOCUMENT {
			DrawEditor()
		} else {
			DrawPrompt(getPrompt())
		}
	}
}
