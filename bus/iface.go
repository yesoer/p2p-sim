package bus

/* To avoid import cycles this file defines all application specific
* event types that may be published, aswell as their embedded data structures.
* Helpful guidelines for naming :
* - Event types should usually be named direction agnostic, so no prefixes like
*   "GUINodeCntChangeEvt". E.g. if they reflect the change of some GUI input
*   name them something like "NodeCountChangeEvt" and if its published by the
*   network, "ResizedNetworkEvt"
* - They get the postfix "Evt" to emphasize their assoziation to the bus.
* - The data structures names should describe the abstract information they carry.
    They do not need to be named with relation to their events or the entire
	eventbus because they might be useful in other, unrelated places aswell.
*/

const NodeDataChangeEvt EventType = "node-data-change"

type NodeData struct {
	TargetId int
	Data     interface{}
}

const ConnectNodesEvt EventType = "connect-nodes"
const DisconnectNodesEvt EventType = "disconnect-nodes"

type Connection struct {
	From int
	To   int
}

const StartNodesEvt EventType = "start-nodes"
const StopNodesEvt EventType = "stop-nodes"

const CodeChangeEvt EventType = "code-change"

type Code string

const NodeCntChangeEvt EventType = "node-count-change"

type NodeCnt int

const NetworkConnectionsEvt EventType = "network-connections"

type Connections []Connection

const NetworkResizeEvt EventType = "network-resize"

type NetworkResize struct {
	Connections
	Cnt int
}

const NodeOutputLogEvt EventType = "node-output-log"

type NodeLog struct {
	Str    string
	NodeId int
}
