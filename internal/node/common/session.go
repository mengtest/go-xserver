package nodecommon

import (
	"github.com/fananchong/go-xserver/common"
	"github.com/fananchong/go-xserver/internal/protocol"
	"github.com/fananchong/gotcp"
)

// ISessionDerived : SessionBase 派生类接口定义
type ISessionDerived interface {
	DoVerify(msg *protocol.MSG_MGR_REGISTER_SERVER, data []byte, flag byte)
	DoRegister(msg *protocol.MSG_MGR_REGISTER_SERVER, data []byte, flag byte)
	DoLose(msg *protocol.MSG_MGR_LOSE_SERVER, data []byte, flag byte)
	DoClose(sessbase *SessionBase)
}

// SessionBase : 网络会话类
type SessionBase struct {
	DefaultNodeInterfaceImpl
	*gotcp.Session
	Ctx     *common.Context
	MsgData []byte
	MsgFlag byte
	derived ISessionDerived
}

// NewSessionBase : 网络会话类的构造函数
func NewSessionBase(ctx *common.Context, derived ISessionDerived) *SessionBase {
	return &SessionBase{
		Ctx:     ctx,
		Session: &gotcp.Session{},
		derived: derived,
	}
}

// OnRecv : 接收到网络数据包，被触发
func (sessbase *SessionBase) OnRecv(data []byte, flag byte) {
	cmd := protocol.CMD_MGR_ENUM(gotcp.GetCmd(data))
	if sessbase.IsVerified() == false && sessbase.doVerify(cmd, data, flag) == false {
		return
	}
	switch cmd {
	case protocol.CMD_MGR_REGISTER_SERVER:
		sessbase.doRegister(data, flag)
	case protocol.CMD_MGR_LOSE_SERVER:
		sessbase.doLose(data, flag)
	case protocol.CMD_MGR_PING:
		// do nothing
	default:
		sessbase.Ctx.Log.Errorln("unknow cmd, cmd =", cmd)
	}
}

// OnClose : 断开连接，被触发
func (sessbase *SessionBase) OnClose() {
	if sessbase.Info != nil {
		XSESSIONMGR.Lose(sessbase)
	}
	sessbase.derived.DoClose(sessbase)
}

// ResetTCPSession : 重置 TCP Session 对象
func (sessbase *SessionBase) ResetTCPSession() {
	sessbase.Session = &gotcp.Session{}
}

func (sessbase *SessionBase) doVerify(cmd protocol.CMD_MGR_ENUM, data []byte, flag byte) bool {
	if cmd == protocol.CMD_MGR_REGISTER_SERVER {
		msg := &protocol.MSG_MGR_REGISTER_SERVER{}
		if gotcp.DecodeCmd(data, flag, msg) == nil {
			sessbase.Close()
			return false
		}
		if msg.GetToken() != sessbase.Ctx.Config.Common.IntranetToken {
			sessbase.Ctx.Log.Errorln("IntranetToken error!",
				"Msg token:", msg.GetToken(),
				"Expect token:", sessbase.Ctx.Config.Common.IntranetToken)
			sessbase.Close()
			return false
		}
		sessbase.derived.DoVerify(msg, data, flag)
		sessbase.Verify()
		return true
	}
	sessbase.Ctx.Log.Errorln("Before message[CMD_MGR_REGISTER_SERVER], recv cmd:", cmd)
	sessbase.Close()
	return false
}

func (sessbase *SessionBase) doRegister(data []byte, flag byte) {
	msg := &protocol.MSG_MGR_REGISTER_SERVER{}
	if gotcp.DecodeCmd(data, flag, msg) == nil {
		sessbase.Close()
		return
	}
	sessbase.derived.DoRegister(msg, data, flag)
}

func (sessbase *SessionBase) doLose(data []byte, flag byte) {
	msg := &protocol.MSG_MGR_LOSE_SERVER{}
	if gotcp.DecodeCmd(data, flag, msg) == nil {
		return
	}
	sessbase.derived.DoLose(msg, data, flag)
}
