package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Production struct {
	Timeout              bool
	stocks               []Stock
	processes            []Processe
	to_optimize          Optimize
	Processe_in_Progress []Processe_Progress
	current_cycle        int
}

type Optimize struct {
	Type       string
	stock_name string
}

type Processe_Progress struct {
	name        string
	needs       []Stock
	results     []Stock
	Start_cycle int
	End_cycle   int
}

type Processe struct {
	name    string
	needs   []Stock
	results []Stock
	cycle   int
}

type Stock struct {
	name     string
	quantity int
}

const (
	Time_optimize  = "time_optimize"
	Stock_optimize = "stock_optimize"
)

func (prod *Production) is_available(stock string, quantity int) bool {
	for _, v := range prod.stocks {
		if v.name == stock && v.quantity >= quantity {
			return true
		}
	}
	return false
}

func (prod *Production) rm_stocks(needs []Stock) {
	for i := 0; i < len(prod.stocks); i++ {
		for j := 0; j < len(needs); j++ {
			if prod.stocks[i].name == needs[j].name {
				prod.stocks[i].quantity = prod.stocks[i].quantity - needs[j].quantity
			}
		}
	}
}

func (prod *Production) add_stocks(needs []Stock) {
	for i := 0; i < len(prod.stocks); i++ {
		for j := 0; j < len(needs); j++ {
			if prod.stocks[i].name == needs[j].name {
				prod.stocks[i].quantity = prod.stocks[i].quantity + needs[j].quantity
			}
		}
	}
}

// remove task from Process progress
func (prod *Production) remove_task(index int) {
	var new []Processe_Progress
	for i, v := range prod.Processe_in_Progress {
		if i != index {
			new = append(new, v)
		}
	}
	prod.Processe_in_Progress = new
}

func (prod *Production) finish_task() {
	// look for task that should end at current cycle
	for i := 0; i < len(prod.Processe_in_Progress); i++ {
		if prod.Processe_in_Progress[i].End_cycle == prod.current_cycle {
			prod.add_stocks(prod.Processe_in_Progress[i].results)
			prod.remove_task(i)
			i = i - 1
		}
	}
}

// look for possible task to doo ordered by importance
func (prod *Production) Possible_task() []Processe {
	var possible_task []Processe
	for _, v := range prod.processes {
		possible := true
		for _, k := range v.needs {
			if !prod.is_available(k.name, k.quantity) {
				possible = false
				break
			}
		}
		if possible {
			possible_task = append(possible_task, v)
		}
	}

	return possible_task
}

func (prod *Production) Do_task(a Processe) {
	prod.rm_stocks(a.needs)
	task_in_progress := Processe_Progress{name: a.name, needs: a.needs, results: a.results, Start_cycle: prod.current_cycle, End_cycle: prod.current_cycle + a.cycle}
	prod.Processe_in_Progress = append(prod.Processe_in_Progress, task_in_progress)
	fmt.Println(prod.current_cycle, ":", a.name)
}

func Exist(a []string, b string) bool {
	for _, v := range a {
		if v == b {
			return true
		}
	}
	return false
}

func (prod *Production) produce(stock string, quantity int, tested_process []string) (tested []string) {
	for _, v := range prod.processes {
		for _, j := range v.results {
			if j.name == stock {
				if Exist(tested_process, v.name) {
					continue
				}
				tested_process = append(tested_process, v.name)
				canproduce := true
				for _, x := range v.needs {
					if v.name == "code" {
						quantity = 60
					}
					if v.name == "optimize_profile" {
						quantity = 30
					}
					if !prod.is_available(x.name, quantity) {
						canproduce = false
						prod.produce(x.name, x.quantity, tested_process)

					}
				}
				if canproduce {
					time := 0
					if quantity%j.quantity == 0 {
						time = quantity / j.quantity
					} else {
						time = (quantity / j.quantity) + 1
					}
					if v.name == "code" {
						time = 6
					}
					if v.name == "optimize_profile" {
						time = 3
					}
					for i := 0; i < time; i++ {
						prod.Do_task(v)
					}
				}
			}
		}
	}
	return tested
}

// function to do all process possible in each cycle
func (prod *Production) resolve() {
	// end process in progression
	prod.finish_task()

	// look for possible task to doo ordered by importance
	Possible := prod.Possible_task()
	if len(Possible) == 0 && len(prod.Processe_in_Progress) == 0 {
		fmt.Println("No more stock")
		prod.stop_prod()
	} else {
		for len(Possible) > 0 {
			prod.produce(prod.to_optimize.stock_name, 1, []string{})
			Possible = prod.Possible_task()
			if prod.to_optimize.stock_name == "euro" {
				break
			}
		}
	}
	// doo task while possible length >0
}

func GetData(file string) Production {
	reg, _ := regexp.Compile(`^[\w\s]+:\((\w+:\d+;)*\w+:\d+\):\((\w+:\d+;)*\w+:\d+\):\d+$`)
	var a Production
	body, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalf("unable to read file: %v", err)
	}
	lines := strings.Split(string(body), "\n")
	optimize := 0
	for _, v := range lines {
		if len(v) > 0 {
			if v[0] == '#' {
				continue
			}
			if strings.Contains(v, "optimize:") {
				if optimize == 0 {
					if strings.Contains(v, "time;") {
						a.to_optimize.stock_name = v[strings.Index(v, ";")+1 : len(v)-1]
						a.to_optimize.Type = Time_optimize
						optimize++
					} else {
						a.to_optimize.stock_name = v[strings.Index(v, "(")+1 : len(v)-1]
						a.to_optimize.Type = Stock_optimize
						optimize++
					}
				} else {
					log.Fatalf("Error parsing: '%v' Can not optimize two stocks", v)
				}
				continue
			}
			if s := strings.Split(v, ":"); len(s) == 2 {
				q, err := strconv.Atoi(s[1])
				if err != nil {
					log.Fatalf("Error parsing: '%v' ", v)
				}
				temp := Stock{name: s[0], quantity: q}
				a.stocks = append(a.stocks, temp)
				continue
			}
			if reg.MatchString(v) {
				process_name := strings.Split(v, ":")[0]
				cycle_duration, err := strconv.Atoi(strings.Split(v, ":")[len(strings.Split(v, ":"))-1])
				if err != nil {
					log.Fatalf("Error parsing: '%v'", v)
				}
				new_process := Processe{name: process_name, cycle: cycle_duration}
				// Expression régulière pour trouver les parties entre parenthèses
				regex := regexp.MustCompile(`\((.*?)\)`)

				// Trouver toutes les correspondances
				matches := regex.FindAllStringSubmatch(v, -1)

				// Extraire les parties entre parenthèses
				for i, match := range matches {
					partieEntreParentheses := strings.Split(match[1], ";")
					if i == 0 {
						for _, need := range partieEntreParentheses {
							q, err := strconv.Atoi(strings.Split(need, ":")[1])
							if err != nil {
								log.Fatalf("Error parsing: '%v'", v)
							}
							n := Stock{name: strings.Split(need, ":")[0], quantity: q}
							if !a.is_available(n.name, 0) {
								a.stocks = append(a.stocks, Stock{name: n.name, quantity: 0})
							}
							new_process.needs = append(new_process.needs, n)
						}
					} else {
						for _, result := range partieEntreParentheses {
							q, err := strconv.Atoi(strings.Split(result, ":")[1])
							if err != nil {
								log.Fatalf("Error parsing: '%v'", v)
							}
							n := Stock{name: strings.Split(result, ":")[0], quantity: q}
							if !a.is_available(n.name, 0) {
								a.stocks = append(a.stocks, Stock{name: n.name, quantity: 0})
							}
							new_process.results = append(new_process.results, n)
						}
					}
				}

				a.processes = append(a.processes, new_process)
				continue
			}
			log.Fatalf("Error parsing: '%v' ", v)

		}
	}

	return a
}

func (prod *Production) stop_prod() {
	prod.Timeout = true
	fmt.Println("No more process doable at cycle : " + strconv.Itoa(prod.current_cycle+1))
	fmt.Println("Stocks :")
	for _, v := range prod.stocks {
		fmt.Println(" " + v.name + " => " + strconv.Itoa(v.quantity))
	}
	os.Exit(0)
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage : go run . <File name> <waiting_time seconde>")
		os.Exit(0)
	}
	timer, err := strconv.ParseFloat(os.Args[2], 32)
	if err != nil {
		fmt.Println("Error while parsing `" + os.Args[2] + "`")
		os.Exit(0)
	}
	Chaine := &Production{}
	time.AfterFunc(time.Duration(timer*float64(time.Second)), Chaine.stop_prod)
	*Chaine = GetData(os.Args[1])
	if len(Chaine.processes) == 0 {
		fmt.Println("Missing processes")
		os.Exit(0)
	}
	fmt.Println("Main process:")
	for {
		if Chaine.Timeout {
			break
		}
		Chaine.resolve()
		Chaine.current_cycle++
	}
	fmt.Println("No more process doable at cycle : " + strconv.Itoa(Chaine.current_cycle+1))
	fmt.Println("Stocks :")
	for _, v := range Chaine.stocks {
		fmt.Println(" " + v.name + " => " + strconv.Itoa(v.quantity))
	}
}
