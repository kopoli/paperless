package paperless

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"
)

/// Generic functionality
func Checksum(data []byte) string {
	return fmt.Sprintf("sha1:%x", sha1.Sum(data))
}

func ChecksumFile(path string) (sum string, err error) {
	var data []byte
	data, err = ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return Checksum(data), err
}

// MkdirParents creates all parent directories of the given path or returns an
// error if they couldn't be created
func MkdirParents(filename string) error {
	dirs := path.Dir(filename)
	return os.MkdirAll(dirs, 0755)
}

/// Parallel processing

type Runner struct {
	RunnerCount int
	ChanBuffer  int
	Job         func(string)
	Finalize    func(string)
}

func CreateRunner(runners int) Runner {
	return Runner{
		RunnerCount: runners,
		ChanBuffer:  10,
		Finalize:    func(string) {},
	}
}

func (run Runner) Do(data []string) {
	var wg sync.WaitGroup
	input := make(chan string, run.ChanBuffer)

	wg.Add(run.RunnerCount)
	for i := 0; i < run.RunnerCount; i++ {
		go func() {
			defer wg.Done()
			for item := range input {
				func() {
					defer run.Finalize(item)
					run.Job(item)
				}()
			}
		}()
	}

	for _, item := range data {
		input <- item
	}
	close(input)
	wg.Wait()

}

type RunnerPool struct {
	runnerCount int
	jobChan     chan Job
	jobChanSize int
	wait        sync.WaitGroup
}

type Job struct {
	Job      func()
	Finalize func()
}

func CreatePool(runners int) RunnerPool {
	ret := RunnerPool{
		runnerCount: runners,
		jobChanSize: 10,
		jobChan:     make(chan Job),
	}

	ret.wait.Add(runners)
	for i := 0; i < runners; i++ {
		go func() {
			defer ret.wait.Done()
			for item := range ret.jobChan {
				func() {
					defer item.Finalize()
					item.Job()
				}()
			}
		}()
	}
	return ret
}

func (pool *RunnerPool) Delete() {
	close(pool.jobChan)
}

func (pool *RunnerPool) Do(job Job) {
	pool.jobChan <- job
}

var Pool *RunnerPool

func CreateDefaultPool(runners int) {
	pool := CreatePool(runners)
	Pool = &pool
}
