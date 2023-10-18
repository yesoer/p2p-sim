package bus

// To avoid import cycles this file defines all events published aswell as their
// data structures.

//---------------------
// PUBLISHED BY GUI

// when node specific data has been changed
var NodeDataChangeEvt EventType = "node-data-changed"

type NodeDataChangeData struct {
	TargetId int
	Data     interface{}
}

// on (dis-)connects
var ConnectNodesEvt EventType = "connect-nodes"
var DisconnectNodesEvt EventType = "disconnect-nodes"

type CheckboxPos struct {
	Ccol int
	Crow int
}

// on signals
var StartEvt EventType = "start"
var StopEvt EventType = "stop"

// on code changes by the editor
var CodeChangedEvt EventType = "code-changed"

var GUINodeCntChangeEvt EventType = "gui-node-count-change"

//---------------------
// PUBLISHED BY BACKEND

// on (dis-)connects
var ConnectionChangeEvt EventType = "connection-changed"

var NetworkNodeCntChangeEvt EventType = "network-node-count-change"
