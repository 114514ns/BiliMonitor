package main

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

type SlaverManager struct {
	Nodes []SlaverNode
	Lock  sync.Mutex
	OnErr func(tasks []string)
}

type SlaverNode struct {
	Address string
	Alive   bool
	Tasks   []string
}

func (man *SlaverManager) AddTask(task string) {
	man.Lock.Lock()
	defer man.Lock.Unlock()

	var aliveNodes []*SlaverNode
	for i := range man.Nodes {
		if man.Nodes[i].Alive && len(man.Nodes[i].Tasks) < 20 {
			aliveNodes = append(aliveNodes, &man.Nodes[i])
		}
	}
	if len(aliveNodes) == 0 {
		if man.OnErr != nil {
			man.OnErr([]string{task})
		}
		return
	}

	target := aliveNodes[rand.Intn(len(aliveNodes))]
	target.Tasks = append(target.Tasks, task)
	go func(addr string) {
		_, err := client.R().Get(addr + "/trace?room=" + task)
		if err != nil && man.OnErr != nil {
			man.OnErr([]string{task})
		}
	}(target.Address)
}

func NewSlaverManager(node []string) *SlaverManager {
	var man = &SlaverManager{}
	for _, s := range node {
		res, err := client.R().Get(s + "/ping")
		if err == nil && res.String() == "pong" {
			man.Nodes = append(man.Nodes, SlaverNode{
				Address: s,
				Alive:   true,
				Tasks:   []string{},
			})
		} else {
			log.Printf("[%s] Error connect to node", s)
		}
	}

	go func() {
		for {
			time.Sleep(10 * time.Second)
			man.Lock.Lock()
			var recoveredTasks []string
			for i := range man.Nodes {
				res, err := client.R().Get(man.Nodes[i].Address + "/ping")
				if err != nil || res.String() != "pong" {
					if man.Nodes[i].Alive {
						log.Printf("[%s] Node down. Reassigning %d tasks", man.Nodes[i].Address, len(man.Nodes[i].Tasks))
						recoveredTasks = append(recoveredTasks, man.Nodes[i].Tasks...)
					}
					man.Nodes[i].Alive = false
					man.Nodes[i].Tasks = nil
				} else {
					man.Nodes[i].Alive = true
				}
			}
			man.Lock.Unlock()

			for _, task := range recoveredTasks {
				man.AddTask(task)
			}
		}
	}()

	return man
}
