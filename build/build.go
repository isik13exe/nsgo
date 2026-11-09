package build

import (
	"fmt"
	"nsgo/chart"
	"nsgo/flags"
	"nsgo/msg"
	"nsgo/tools"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

type NSRun struct {
	Delay      int
	Bandwidth  int
	NPackte    int
	PacketSize int
	ConfigFile string
	filename   string
	outfile    string
	Pcapdir    string
}

type Testholder struct {
	Path       string
	Name       string
	Delay      []int
	Bandwidth  []int
	NPacket    []int
	PacketSize []int
}

type Testcase struct {
	Delay      int
	Bandwidth  int
	NPacket    int
	PacketSize int
}

func (n NSRun) String() string {
	return fmt.Sprintf("NS{\n\tDelay = %d \n\tBandwidth = %d \n\tNPackte = %d \n\tPacketSize = %d \n\tConfigFile = %s \n\tPcapdir = %s \n}", n.Delay, n.Bandwidth, n.NPackte, n.PacketSize, n.ConfigFile, n.Pcapdir)
}

var NS3 NSRun
var conf []byte
var Success string = "build and run succeeded"
var Testfile string
var Testc Testholder
var Testcases []Testcase
var Testing bool

func Init() tools.Error {
	var holder NSRun
	flags.DescTool("nsgo is a go tool to run ns3 files and retreive the results:\n\t nsgo <file> [options]\n")
	flags.StringVar(&NS3.ConfigFile, "conf", "", "feed a config file to the program")
	flags.IntVar(&holder.Bandwidth, "band", 0, "set the bandwidth in MB")
	flags.IntVar(&holder.Delay, "delay", 0, "set the delay in ms")
	flags.IntVar(&holder.NPackte, "npackets", 0, "set the number of packets to send")
	flags.IntVar(&holder.PacketSize, "packetsize", 0, "set the size of packets to send")
	flags.StringVar(&holder.Pcapdir, "pcapdir", "", "set the directory to save pcap files")
	flags.StringVar(&Testfile, "test", "", "stress test with a bunch of data")

	err := flags.Catch()
	help := flags.UseHelp()
	if help != nil {
		return help
	}
	if err != nil {
		return tools.NewErrorf(err.Code(), "Parsing flags: %v", err)
	}
	if Testfile != "" {
		if err := parseTest(Testfile); err != nil {
			return err
		}
		Testing = true
		NS3.filename = flags.Args[0]
		return nil
	}
	if NS3.ConfigFile != "" {
		var errs error
		conf, errs = os.ReadFile(NS3.ConfigFile)
		if errs != nil {
			return tools.NewErrorf(tools.ErrBuildGeneral, "could not read config file: %w", errs)
		}
		errs = NS3.FillConf()
		if errs != nil {
			return tools.NewErrorf(tools.ErrBuildGeneral, "could not fill config: %w", errs)
		}
	} else {
		if holder.Bandwidth == 0 || holder.Delay == 0 || holder.NPackte == 0 || holder.PacketSize == 0 {
			return tools.NewErrorf(tools.ErrBuildGeneral, "no config file or data provided.")
		}
	}
	if holder.Bandwidth != 0 {
		NS3.Bandwidth = holder.Bandwidth
	}
	if holder.Delay != 0 {
		NS3.Delay = holder.Delay
	}
	if holder.NPackte != 0 {
		NS3.NPackte = holder.NPackte
	}
	if holder.PacketSize != 0 {
		NS3.PacketSize = holder.PacketSize
	}
	if NS3.filename == "" {
		NS3.filename = flags.Args[0]
	}
	if holder.Pcapdir != "" {
		NS3.Pcapdir = holder.Pcapdir
	}
	if NS3.Pcapdir != "" {
		NS3.Pcapdir = NS3.Pcapdir
	}
	if err := tools.PurgeDir(NS3.Pcapdir); err != nil {
		return tools.NewErrorf(err.Code(), "could not purge pcap directory: %w", err)
	}
	if err := tools.Mkdir(NS3.Pcapdir); err != nil {
		return tools.NewErrorf(err.Code(), "could not create pcap directory: %w", err)
	}
	var errs error
	NS3.Pcapdir, errs = filepath.Abs(NS3.Pcapdir)
	if errs != nil {
		return tools.NewErrorf(tools.ErrBuildGeneral, "could not get absolute path for pcap directory: %w", errs)
	}
	return nil
}

func parseTest(filename string) tools.Error {
	rawdata, err := os.ReadFile(filename)
	if err != nil {
		return tools.NewErrorf(tools.ErrBuildGeneral, "could not read test file: %w", err)
	}
	done := map[string]bool{
		"delay":      false,
		"bandwidth":  false,
		"npacket":    false,
		"name":       false,
		"packetsize": false,
	}
	lines := tools.Skipnl(string(rawdata))
	for i, raw := range lines {
		cas, data, err := tools.Treatline(raw)
		if err != nil {
			return tools.NewErrorf(tools.ErrBuildGeneral, "Error when parsing the config file:\n\t line:%d   %v", i, err)
		}
		_, ok := done[cas]
		if !ok {
			return tools.NewErrorf(tools.ErrBuildGeneral, "Error when parsing the test config file:\n\t line:%d   parsing %s resulted in an error.\n\t\t Unknown config paramter", i, cas)
		}
		if done[cas] {
			return tools.NewErrorf(tools.ErrBuildGeneral, "Error when parsing the test config file:\n\t line:%d   parsing %s resulted in an error.\n\t\t confi paramter present twice ", i, cas)
		}
		done[cas] = true
		switch cas {
		case "packetsize":
			hold, err := tools.ParseTest(data)
			if err != nil {
				return tools.NewErrorf(tools.ErrBuildGeneral, "Error when parsing the test config file:\n\t line:%d   parsing %s resulted in an error.\n\t\t %v", i, cas, err)
			}
			Testc.PacketSize = hold
		case "delay":
			hold, err := tools.ParseTest(data)
			if err != nil {
				return tools.NewErrorf(tools.ErrBuildGeneral, "Error when parsing the test config file:\n\t line:%d   parsing %s resulted in an error.\n\t\t %v", i, cas, err)
			}
			Testc.Delay = hold
		case "bandwidth":
			hold, err := tools.ParseTest(data)
			if err != nil {
				return tools.NewErrorf(tools.ErrBuildGeneral, "Error when parsing the test config file:\n\t line:%d   parsing %s resulted in an error.\n\t\t %v", i, cas, err)
			}
			Testc.Bandwidth = hold
		case "npacket":
			hold, err := tools.ParseTest(data)
			if err != nil {
				return tools.NewErrorf(tools.ErrBuildGeneral, "Error when parsing the test config file:\n\t line:%d   parsing %s resulted in an error.\n\t\t %v", i, cas, err)
			}
			Testc.NPacket = hold
		case "name":
			Testc.Name = data
		default:
			return tools.NewErrorf(tools.ErrBuildGeneral, "Error when parsing the test config file:\n\t line:%d   parsing %s resulted in an error.\n\t\t Unknown config paramter", i, cas)
		}
	}
	for k, v := range done {
		if !v {
			return tools.NewErrorf(tools.ErrBuildGeneral, "Error when parsing the test config file:\n\t\t %s is required", k)
		}
	}
	return nil
}

func GenerateTestcases() []Testcase {
	for _, delay := range Testc.Delay {
		for _, bandwidth := range Testc.Bandwidth {
			for _, nPacket := range Testc.NPacket {
				for _, packetSize := range Testc.PacketSize {
					Testcases = append(Testcases, Testcase{
						Delay:      delay,
						Bandwidth:  bandwidth,
						NPacket:    nPacket,
						PacketSize: packetSize,
					})
				}
			}
		}
	}
	return Testcases
}

func (n *NSRun) FillConf() tools.Error {
	lines := tools.Skipnl(string(conf))

	done := map[string]bool{
		"delay":      false,
		"bandwidth":  false,
		"npacket":    false,
		"pcapdir":    false,
		"packetsize": false,
	}

	for i, raw := range lines {
		cas, data, err := tools.Treatline(raw)
		if err != nil {
			return tools.NewErrorf(err.Code(), "Error when parsing the config file:\n\t line:%d   %v", i, err)
		}
		_, ok := done[cas]
		if !ok {
			return tools.NewErrorf(tools.ErrBuildGeneral, "Error when parsing the config file:\n\t line:%d   parsing \"%s\" resulted in an error.\n\t\t Unknown config paramter", i, cas)
		}
		if done[cas] {
			return tools.NewErrorf(tools.ErrBuildGeneral, "Error when parsing the config file:\n\t line:%d   parsing \"%s\" resulted in an error.\n\t\t Config paramter present twice ", i, cas)
		}
		done[cas] = true
		switch cas {
		case "packetsize":
			var tmp int
			var errs error
			tmp, errs = strconv.Atoi(data)
			if errs != nil {
				return tools.NewErrorf(tools.ErrBuildGeneral, "Error when parsing the config file:\n\t line:%d   parsing \"%s\" resulted in an error.", i, cas, errs)
			}
			n.PacketSize = tmp
		case "pcapdir":
			n.Pcapdir = data
		case "outfile":
			n.outfile = data
		case "file":
			n.filename = data
		case "delay":
			var tmp int
			var errs error
			tmp, errs = strconv.Atoi(data)
			if errs != nil {
				return tools.NewErrorf(tools.ErrBuildGeneral, "Error when parsing the config file:\n\t line:%d   parsing \"%s\" resulted in an error.", i, cas, errs)
			}
			n.Delay = tmp
		case "bandwidth":
			var tmp int
			var errs error
			tmp, errs = strconv.Atoi(data)
			if errs != nil {
				return tools.NewErrorf(tools.ErrBuildGeneral, "Error when parsing the config file:\n\t line:%d   parsing \"%s\" resulted in an error.", i, cas, errs)
			}
			n.Bandwidth = tmp
		case "npacket":
			var tmp int
			var errs error
			tmp, errs = strconv.Atoi(data)
			if errs != nil {
				return tools.NewErrorf(tools.ErrBuildGeneral, "Error when parsing the config file:\n\t line:%d   parsing \"%s\" resulted in an error.", i, cas, errs)
			}
			n.NPackte = tmp
		}
	}
	for k, v := range done {
		if !v {
			return tools.NewErrorf(tools.ErrBuildGeneral, "Error when parsing the config file:\n\t\t %s is required", k)
		}
	}
	return nil
}

func (n NSRun) load() []string {

	return []string{
		"run",
		fmt.Sprintf("%s --np=%d --bn=%d --delay=%d --pcapdir=%s --ps=%d", n.filename, n.NPackte, n.Bandwidth, n.Delay, n.Pcapdir, n.PacketSize),
	}
}

func Build() tools.Error {
	cmd := exec.Command("ns3", NS3.load()...)
	es := msg.ErrorSplitter()
	fs := msg.FileSplitter(NS3.outfile)
	cmd.Stderr = es
	cmd.Stdout = fs
	msg.INFOln("Executing", cmd.Args)
	if err := cmd.Run(); err != nil {
		return tools.NewErrorf(tools.ErrBuildGeneral, "%s", err)
	}
	tr := NS3.Pcapdir + "/" + fmt.Sprintf("tcp-np%d-bn%d-d%d.tr", NS3.NPackte, NS3.Bandwidth, NS3.Delay)
	msg.INFOln("Rendering charts from", tr)
	if err := chart.Graph(tr, NS3.Pcapdir); err != nil {
		return err
	}
	return nil
}

func Prepare_dir() tools.Error {
	if err := tools.PurgeDir(Testc.Name); err != nil {
		return tools.NewErrorf(err.Code(), "could not purge pcap directory: %w", err)
	}
	if err := tools.Mkdir(Testc.Name); err != nil {
		return tools.NewErrorf(err.Code(), "could not create pcap directory: %w", err)
	}
	var err error
	Testc.Path, err = filepath.Abs(Testc.Name)
	if err != nil {
		return tools.NewErrorf(tools.ErrBuildGeneral, "could not get absolute path: %w", err)
	}
	for _, test := range Testcases {
		if err := tools.Mkdir(Testc.Path + "/" + fmt.Sprintf("%d-%d-%d", test.NPacket, test.Bandwidth, test.Delay)); err != nil {
			return tools.NewErrorf(err.Code(), "could not create test directory: %w", err)
		}
	}
	return nil
}

func Tests() tools.Error {
	GenerateTestcases()
	if err := Prepare_dir(); err != nil {
		return err
	}
	msg.INFOln("starting tests", Testcases)
	for _, test := range Testcases {
		msg.INFOln("testing", test)
		NS3.NPackte = test.NPacket
		NS3.Bandwidth = test.Bandwidth
		NS3.Delay = test.Delay
		NS3.Pcapdir = Testc.Path + "/" + fmt.Sprintf("%d-%d-%d", test.NPacket, test.Bandwidth, test.Delay)
		NS3.PacketSize = test.PacketSize
		if err := Build(); err != nil {
			return err
		}
	}
	return nil
}
