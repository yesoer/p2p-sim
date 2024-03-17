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
  - Declare wrappers even for simple data e.g. NodeId for int. Since the parameter
    types in Bind callbacks cannot be checked by the compiler, switching from int
	to a more complex type later on would not be caught and refactoring becomes
	increasingly complex and error prone. By wrapping types the parameter will
	always be correct and changes will cause errors in the callbacks body which
	are then caught by the compiler.
*/

const NodeDataChangeEvt EventType = "node-data-change"

type NodeData struct {
	TargetId int
	Data     any
}

const ConnectNodesEvt EventType = "connect-nodes"
const DisconnectNodesEvt EventType = "disconnect-nodes"

type Connection struct {
	From int
	To   int
}

const StartNodesEvt EventType = "start-nodes"
const StopNodesEvt EventType = "stop-nodes"
const DebugNodesEvt EventType = "debug-nodes"
const ContinueNodesEvt EventType = "continue-nodes"

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

const NodeOutputEvt EventType = "node-output"

type NodeOutput struct {
	Log    string
	Result any
	NodeId int
}

const SentToEvt EventType = "sent-to"

type SendTask struct {
	From int
	To   int
	Data any
}

const AwaitStartEvt EventType = "await-start"
const AwaitEndEvt EventType = "await-end"

type NodeId int // TODO : if we keep this, other structs should use it aswell

const ProjectOpenEvt EventType = "project-open"

const FileOpenEvt EventType = "file-open"

type Source string

const (
	LocalFile    Source = "local-file"
	EmbeddedFile Source = "embedded-file"
	LocalDir     Source = "local-dir"
)

type File struct {
	Path   string
	Source Source
}
