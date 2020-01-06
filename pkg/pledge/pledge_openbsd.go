package pledge

import "golang.org/x/sys/unix"

func Pledge(promises, execpromises string) error {
    return unix.Pledge(promises, execpromises)
}

func PledgePromises(promises string) error {
    return unix.PledgePromises(promises)
}

func PledgeExecpromises(execpromises string) error {
    return unix.PledgeExecpromises(execpromises)
}
