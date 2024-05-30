package bus

/* To avoid import cycles this file defines all application specific
* event types that may be published, aswell as their embedded data structures.
* Helpful guidelines for naming :
* - Event types should usually be named direction agnostic, so no prefixes like
*   "GUINodeCntChangeEvt". E.g. if they reflect the change of some GUI input
*   name them something like "NodeCountChangeEvt" and if its published by the
*   network, "ResizedNetworkEvt" (this because GUI components may influence
*   each other)
* - They get the postfix "Evt" to emphasize their assoziation to the bus.
* - The data structures names should describe the abstract information they carry.
*   They do not need to be named with relation to their events or the entire
*   eventbus because they might be useful in other, unrelated places aswell.
* - Declare wrappers even for simple data e.g. NodeId for int. Since the parameter
*   types in Bind callbacks cannot be checked by the compiler, switching from int
*   to a more complex type later on would not be caught and refactoring becomes
*   increasingly complex and error prone. By wrapping types the parameter will
*   always be correct and changes will cause errors in the callbacks body which
*   are then caught by the compiler.
 */

// When a nodes custom input data changes, this event is published
const NodeDataChangeEvt EventType = "node-data-change"

type NodeData struct {
	TargetId int
	Data     any
}

// When nodes get (dis-)connected (usually through the GUI) these events are
// published
const ConnectNodesEvt EventType = "connect-nodes"
const DisconnectNodesEvt EventType = "disconnect-nodes"

type Connection struct {
	From int
	To   int
}

// The following events transfer the nodes from one state to another
const StartNodesEvt EventType = "start-nodes"
const StopNodesEvt EventType = "stop-nodes"
const DebugNodesEvt EventType = "debug-nodes"
const ContinueNodesEvt EventType = "continue-nodes"

// This event notifies the network of a change in the amount of nodes to be used
const NodeCntChangeEvt EventType = "node-count-change"

type NodeCnt int

// When the network is resized, this event is published to inform the GUI
// TODO : Check if this is necessary or if the GUI can keep track itself from the
// connect/disconnect events. If its necessary as some sort of confirmation,
// a rule is needed to clear up which events need "confirmation" events and make
// them easier to identify e.g. by a pre-/postfix.
const NetworkConnectionsEvt EventType = "network-connections"

type Connections []Connection

// When the network is resized, this event is published to inform the GUI
// TODO : similar to the NetworkConnectionsEvt, this seems to be required for
// confirmation as some resize inputs are invalid (e.g. negative values). The
// GUI should just not allow such values, making this event unnecessary.
const NetworkResizeEvt EventType = "network-resize"

type NetworkResize struct {
	Connections
	Cnt int
}

// This event may be used by nodes to communicate their logs to the GUI
const NodeOutputEvt EventType = "node-output"

type NodeOutput struct {
	Log    string
	Result any
	NodeId int
}

// SentToEvt is used by nodes to communicate to the GUI that they have sent
// data to another node
const SentToEvt EventType = "sent-to"

type SendTask struct {
	From int
	To   int
	Data any
}

// The await events are used by nodes to signal the GUI that they are waiting
// for some message to arrive
const AwaitStartEvt EventType = "await-start"
const AwaitEndEvt EventType = "await-end"

// TODO : if we keep this, other structs should use it aswell, e.g. the
// NodeOutput struct
type NodeId int

// OpenEvt communicates that a project has been opened
const OpenEvt EventType = "open"

type SourceType string

const (
	File      SourceType = "file"
	Directory SourceType = "directory"
)

type Source struct {
	Path string
	Type SourceType
}

// The following events are used to communicate the selection of an editor
const EditorSelectEvt EventType = "editor-select"

type EditorType string

const (
	Neovim EditorType = "neovim"
	Entry  EditorType = "entry"
)
