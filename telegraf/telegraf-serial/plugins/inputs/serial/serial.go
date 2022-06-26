package serial
// serial.go

import (
	"strconv"
	"strings"
	"time"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"go.bug.st/serial"
)

type Serial struct {
	Parity		string	`toml:"parity"`
	BaudRate  int             `toml:"baudrate"`
	DataBits	int	`toml:"databits"`
	StopBits	string	`toml:"stopbits"`
	RTS	bool	`toml:"rts"`
	DTR	bool	`toml:"dtr"`
	Log telegraf.Logger	`toml:"-"`

	localParity	serial.Parity
	localStopBits	serial.StopBits
	isConnected	bool
	port	serial.Port
}


func (s *Serial) Description() string {
	return "a serial plugin"
}

func (s *Serial) SampleConfig() string {
	return `
  ## Indicate if everything is fine
  ok = true
`
}

// Init is for setup, and validating config.
func (s *Serial) Init() error {
	err := s.initConn()
	if err != nil {
		s.Log.Debugf("Cannot init serial connection (maybe wrong configuration)");
		return err
	}
	s.connect()
	return nil
}

func (s *Serial) Gather(acc telegraf.Accumulator) error {
	if (s.isConnected == false){
		s.connect()
		return nil
	}

	buff := make([]byte, 100)
	// Reads up to 100 bytes
	n, err := s.port.Read(buff)
	if err != nil {
		s.Log.Debugf("Can't read from serial port",err)
		s.isConnected = false
	}
	index := strings.Index(string(buff[:n]),"000");
	if index != 0 {
		return nil
	}
	temp,_ := strconv.Atoi(string(buff[3:5]))

	fieldsG := map[string]interface{}{
		"temp": temp,
	}
	now := time.Now()
	acc.AddGauge("serial", fieldsG, nil, now)

	return nil
}


func (s *Serial) readConfig () error {

	//check the Parity ENUM

	switch {
		case s.Parity == "N":
			s.localParity = serial.NoParity
		case s.Parity == "O":
			s.localParity = serial.OddParity
		case s.Parity == "E":
			s.localParity = serial.EvenParity
		case s.Parity == "M":
			s.localParity = serial.MarkParity
		case s.Parity == "S":
			s.localParity = serial.SpaceParity
	}
	//check the StopBits ENUM
	switch {
		case s.StopBits == "1":
			s.localStopBits = serial.OneStopBit
		case s.StopBits == "1.5":
			s.localStopBits = serial.OnePointFiveStopBits
		case s.StopBits == "2":
			s.localStopBits = serial.TwoStopBits
	}
	return nil
}

func (s *Serial) connect () error {
	ports, err := serial.GetPortsList()
	if err != nil {
		s.Log.Errorf("Some unexpected error occured. %v",err)
		s.isConnected = false
		return nil
	}
	if len(ports) == 0 {
		s.Log.Warnf("No serial ports found!")
		s.isConnected = false
		return nil
	}
	// Print the list of detected ports
	for _, port := range ports {
		s.Log.Infof("Found port %v\n", port)

	}

	// Open the first serial port detected at 2400bps O71
	mode := &serial.Mode{
		BaudRate: s.BaudRate,
		Parity:   s.localParity,
		DataBits: s.DataBits,
		StopBits: s.localStopBits,
	}
	port, err := serial.Open(ports[0], mode)
	s.port = port
	if err != nil {
		s.isConnected = false
		s.Log.Warnf("I couldn't open the port because: %s ",err.Error())
		return nil
	}
	s.port.ResetInputBuffer()
	if s.DTR == true {
		err1 := s.port.SetDTR(true)
		if err1 != nil {
			s.Log.Debugf("Can't set DTR true",err1)
		}
	} else {
		err1 := s.port.SetDTR(false)
		if err1 != nil {
			s.Log.Debugf("Can't set DTR false",err1)
		}

	}
	if s.RTS == true {
		err1 := s.port.SetRTS(true)
		if err1 != nil {
			s.Log.Debugf("Can't set DTR true",err1)
		}
	} else {
		err1 := s.port.SetRTS(false)
		if err1 != nil {
			s.Log.Debugf("Can't set DTR false",err1)
		}

	}
	status, err := s.port.GetModemStatusBits()
	if err != nil {
		s.Log.Debugf("Can't get serial status",err)
	}
	s.Log.Debugf("Status: %+v\n", status)

	s.isConnected = true
	return nil
}

func (s *Serial) initConn () error {
	s.readConfig()
	s.isConnected = false
	return nil
}

func (s *Serial) Stop() {
	s.port.Close()
}

func init() {
	inputs.Add("serial", func() telegraf.Input {
		return &Serial{}
	})
}

