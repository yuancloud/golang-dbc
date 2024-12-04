package device

import (
	"encoding/binary"
	"fmt"
	"github.com/sirupsen/logrus"
	"golang-dbc/pkg/utils"
	"net"
	"sync"
	"time"
)

type Frame struct {
	RawBytes []byte
	RcvTime  int64
}

type Device struct {
	IP         string
	Port       int
	VarSyncers *sync.Map // "20": VarSyncer1, "21": VarSyncer2
	Conn       net.Conn
	FrameChan  chan *Frame
	DBCDecoder func(int, string) string
	DBCVars    *sync.Map
	DBCContent string
}

func NewDevice(ip string, port int, dbcContent string) (*Device, error) {
	d := &Device{
		IP:         ip,
		Port:       port,
		VarSyncers: &sync.Map{},
		FrameChan:  make(chan *Frame, 100000),
		DBCContent: dbcContent,
		DBCVars:    &sync.Map{},
	}
	var err error
	d.DBCDecoder, err = utils.GetDBCDecoder(dbcContent)
	return d, err
}

func (d *Device) Start() error {
	err := d.Connect(3)
	if err != nil {
		return err
	}
	go d.Receive()
	return nil
}

func (d *Device) Connect(timeout int) error {
	addr := fmt.Sprintf("%s:%d", d.IP, d.Port)
	dialer := net.Dialer{Timeout: time.Duration(timeout) * time.Second}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return err
	}
	d.Conn = conn
	return nil
}

func (d *Device) Receive() {
	for {
		buf := [13 * 8]byte{}
		if d.Conn == nil {
			logrus.Infof("reconnect device %v:%v ...", d.IP, d.Port)
			d.Connect(3)
			continue
		}
		if err := d.Conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
			logrus.Errorf("SetReadDeadline Failed: %v", err)
			continue
		}
		count, err := d.Conn.Read(buf[:])
		if err != nil {
			logrus.Infof("Read err: %v", err)
			logrus.Infof("reconnect device %v:%v ...", d.IP, d.Port)
			d.Connect(3)
			continue
		}
		if count%13 != 0 || count == 0 {
			continue
		}
		for err != nil {
			logrus.Infof("reconnect device %v:%v ...", d.IP, d.Port)
			d.Connect(3)
			time.Sleep(time.Second)
			continue
		}

		for i := 0; i < count; {
			frame := &Frame{
				RawBytes: buf[i : i+13],
				RcvTime:  time.Now().UnixMilli(),
			}
			msgIDBytes := frame.RawBytes[1:5]
			msgDataBytes := frame.RawBytes[5:]
			msgID := binary.BigEndian.Uint32(msgIDBytes) + 2147483648
			msgData := "[%v,%v,%v,%v,%v,%v,%v,%v]"
			if len(msgDataBytes) == 8 {
				msgData = fmt.Sprintf(msgData, msgDataBytes[0], msgDataBytes[1], msgDataBytes[2], msgDataBytes[3],
					msgDataBytes[4], msgDataBytes[5], msgDataBytes[6], msgDataBytes[7])
			}

			res, err := utils.ComputeDBCVars(int(msgID), msgData, d.DBCDecoder)
			if err != nil {
				logrus.Infof("计算dbc信号失败, msgID[%v], msgData[%v]", msgID, msgData)
			}
			for k, v := range res {
				d.DBCVars.Store(k, v)
			}
			i += 13
		}
	}
}
