package msgbroker

type MsgEmitter interface {
	Emit(msg DataMessage) error 
}
