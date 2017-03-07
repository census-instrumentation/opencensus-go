package adapter

import (
	"net"
)

type ConnectionsMgr interface {
	NewClientConnStatus(addr net.Addr) *clientConnStatus
	RemoveClientConnStatus(cs *clientConnStatus)
	NewServerConnStatus(c *counter, localAddr, remoteAddr net.Addr) *serverConnStatus
	RemoveServerConnStatus(cs *serverConnStatus)
}

type connectionsMgr struct{}

func DefaultManager() ConnectionsMgr {
	return &connectionsMgr{}
}

func (cm *connectionsMgr) NewClientConnStatus(addr net.Addr) *clientConnStatus {
	return &clientConnStatus{}
}

func (cm *connectionsMgr) RemoveClientConnStatus(cs *clientConnStatus) {
}

func (cm *connectionsMgr) NewServerConnStatus(c *counter, localAddr, remoteAddr net.Addr) *serverConnStatus {
	return &serverConnStatus{}
}

func (cm *connectionsMgr) RemoveServerConnStatus(cs *serverConnStatus) {
}
