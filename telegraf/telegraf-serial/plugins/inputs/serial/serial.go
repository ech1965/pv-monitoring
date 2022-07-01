package serial

// serial.go

import (
	"bufio"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/parsers/influx"

	"go.bug.st/serial"
)

type Serial struct {
	Parity		string	`toml:"parity"`
	BaudRate    int     `toml:"baudrate"`
	DataBits	int   	`toml:"databits"`
	StopBits	string	`toml:"stopbits"`
	RTS			bool	`toml:"rts"`
	DTR			bool	`toml:"dtr"`
	Device		string  `toml:"device"`
	Log telegraf.Logger	`toml:"-"`


	localParity	serial.Parity
	localStopBits	serial.StopBits
	isConnected	bool
	port	serial.Port
	scanner  *bufio.Scanner
	parser   parsers.Parser
}


func (s *Serial) Description() string {
	return "Read lines from serial port"
}

func (s *Serial) SampleConfig() string {
	return `
	[[inputs.serial]]
	baudrate = 115200
	device = "/dev/ttyUSB0"
    databits = 8
	stopbits = "1"
	parity = "N"
`
}
func (s *Serial) SetParser(p parsers.Parser) {
	s.parser = p 
	s.Log.Debugf("parser set");
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
	
	for s.scanner.Scan() {
		line := s.scanner.Text()
		metrics, err := s.parser.Parse([]byte(line))
        if err == nil {
			for _, m := range metrics {
				acc.AddMetric(m)	
			}
			
		}
	}
	
	err := s.scanner.Err()
	if err != nil {
		s.Log.Debugf("Can't read from serial port",err)
		s.isConnected = false
	}
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
	if s.Device == "" {
		s.Device = "/dev/ttyUSB0"
	}	
	handler := influx.NewMetricHandler()
	parser  := influx.NewParser(handler)
	parser.SetTimeFunc(DefaultTime)
	s.SetParser(parser)
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
	found := false
	for _, port := range ports {
		s.Log.Infof("Found port %v\n", port)
		if port == s.Device {
			found = true
		}
	}
	if !found {
		s.Log.Warnf("Configured serial port not found!=%s",s.Device)
		s.isConnected = false
		return nil
	}

	// Open the first serial port detected at 2400bps O71
	mode := &serial.Mode{
		BaudRate: s.BaudRate,
		Parity:   s.localParity,
		DataBits: s.DataBits,
		StopBits: s.localStopBits,
	}
	port, err := serial.Open(s.Device, mode)
	s.port = port
	s.scanner = bufio.NewScanner(port)

	
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
var DefaultTime = func() time.Time {
	return time.Now()
}

