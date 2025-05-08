package main

import (
	"github.com/bytedance/sonic"
	"log"
	"math/rand"
	"strconv"
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

	for _, node := range man.Nodes {
		for _, t := range node.Tasks {
			if t == task {
				log.Printf("[%s] task is running already", task)
				return
			}
		}
	}
	if Has(config.BlackTracing, task) {
		log.Printf("[%s] skip because task is in black list", task)
		return
	}

	var aliveNodes []*SlaverNode
	for i := range man.Nodes {
		if man.Nodes[i].Alive && len(man.Nodes[i].Tasks) < 25 || Has(config.Tracing, task) /*特别关注列表内无视限制*/ {
			aliveNodes = append(aliveNodes, &man.Nodes[i])
		}
	}
	if len(aliveNodes) == 0 {
		log.Printf("[%s] cannot add because queue is full", task)
		if man.OnErr != nil {
			man.OnErr([]string{task})
		}
		return
	}

	target := aliveNodes[rand.Intn(len(aliveNodes))]
	target.Tasks = append(target.Tasks, task)

	go func(addr string) {
		res, err := client.R().
			Get(addr + "/trace?room=" + task)

		log.Printf("[%s]  responsed with  %d", task, res.StatusCode())
		if err != nil && man.OnErr != nil {
			man.OnErr([]string{task})
		}
	}(target.Address)
}
func (man *SlaverManager) RemoveTasks(tasks []string) {
	man.Lock.Lock()
	defer man.Lock.Unlock()

	// 构建快速查找表
	toRemove := make(map[string]struct{})
	for _, t := range tasks {
		toRemove[t] = struct{}{}
	}

	for i := range man.Nodes {
		var newTasks []string
		for _, task := range man.Nodes[i].Tasks {
			if _, found := toRemove[task]; !found {
				newTasks = append(newTasks, task)
			} else {
				log.Printf("从节点 %s 移除任务 %s", man.Nodes[i].Address, task)
			}
		}
		man.Nodes[i].Tasks = newTasks
	}
}
func (man *SlaverManager) GetAllTasks() []string {
	man.Lock.Lock()
	defer man.Lock.Unlock()

	var result []string
	seen := make(map[string]struct{})

	for _, node := range man.Nodes {
		for _, task := range node.Tasks {
			if _, exists := seen[task]; !exists {
				seen[task] = struct{}{}
				result = append(result, task)
			}
		}
	}
	return result
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
	go func() {
		ticker := time.NewTicker(90 * time.Second)
		defer ticker.Stop()

		for range ticker.C {

			all := man.GetAllTasks()

			for _, s := range all {
				var u = "https://api.live.bilibili.com/xlive/web-room/v1/index/getRoomBaseInfo?req_biz=web_room_componet&room_ids=" + s
				r, err := client.R().Get(u)
				if err != nil {
					log.Printf("[%s] Error get room info %v", s, err)
				}
				var o map[string]interface{}
				sonic.Unmarshal(r.Body(), &o)
				for _, i := range o["data"].(map[string]interface{})["by_room_ids"].(map[string]interface{}) {
					if i.(map[string]interface{})["live_status"].(float64) != 1 {
						if !Has(config.Tracing, s) {
							man.RemoveTasks([]string{s})
						}
					}
				}
				time.Sleep(500 * time.Millisecond)
			}

		}
	}()

	return man
}
func (man *SlaverManager) isSelf(address string) bool {
	return "http://127.0.0.1:"+strconv.Itoa(int(config.Port)) == address
}
func (man *SlaverManager) GetNodeByTask(task string) (string, bool) {
	man.Lock.Lock()
	defer man.Lock.Unlock()

	for _, node := range man.Nodes {
		for _, t := range node.Tasks {
			if t == task {
				return node.Address, true
			}
		}
	}
	return "", false
}
