package sysmon

type Usage struct {
	CPU []CPUCoreUsage
}

type CPUCoreUsage struct {
	Slug       string
	Percentage int64
}

type cpuCoreLoad struct {
	Slug      string
	User      int64 // User is time spent in user mode.
	Nice      int64 // Nice is time spent in user mode with low priority.
	System    int64 // System is time spent in system mode.
	Idle      int64 // Idle is time spent in the idle task.
	IOWait    int64 // IOWait is time waiting for I/O to complete
	IRq       int64 // IRq is time servicing interrupts.
	SoftIRq   int64 // SoftIRq is time servicing softirqs.
	Steal     int64 // Steal is stolen by other operating systems time.
	Guest     int64 // Guest is time spent running a virtual CPU for guest operating systems.
	GuestNice int64 // GuestNice is time spent running a niced guest.
}

func (cpuсl *cpuCoreLoad) GetTotalIdle() int64 {
	return cpuсl.Idle + cpuсl.IOWait
}

func (cpuсl *cpuCoreLoad) GetTotalNonIdle() int64 {
	return cpuсl.User + cpuсl.Nice + cpuсl.System + cpuсl.IRq + cpuсl.SoftIRq + cpuсl.Steal
}

func (cpuсl *cpuCoreLoad) GetTotal() int64 {
	return cpuсl.GetTotalIdle() + cpuсl.GetTotalNonIdle()
}
