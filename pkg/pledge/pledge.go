// +build !openbsd

package pledge

func Pledge(promises, execpromises string) error {
    return nil
}

func PledgePromises(promises string) error {
    return nil
}

func PledgeExecpromises(execpromises string) error {
    return nil
}
