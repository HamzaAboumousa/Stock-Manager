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

func (prod *Production) hwo_produce(stock string) []Processe {
	var result []Processe
	for _, v := range prod.processes {
		for _, k := range v.results {
			if k.name == stock && k.quantity > 0 {
				result = append(result, v)
			}
		}
	}
	return result
}

// func (prod Production) processe_to_do(Stock string) bool {
// 	if Stock != prod.to_optimize.stock_name {
// 		if !prod.is_available(Stock, 1) && !prod.processe_to_do(Stock) {
// 			return false
// 		}
// 	}
// 	for _,v := range prod.processes {
// 		for _,k := range v.results{
// 			if k.name
// 		}
// 	}
// 	return true
// }

func (prod *Production) optimize_by_Time() {
}

func (prod *Production) optimize_by_Stock() {
}

func (prod *Production) resolve() {
	// if !prod.processe_to_do(prod.to_optimize.stock_name) {
	// 	fmt.Println("Error : No process possible")
	// 	os.Exit(0)
	// }
	if prod.to_optimize.Type == Time_optimize {
		prod.optimize_by_Time()
	} else {
		prod.optimize_by_Stock()
	}
}

func GetData(file string) Production {
	reg, _ := regexp.Compile(`^[\w\s]+:\((\w+:\d+;)*\w+:\d+\):\((\w+:\d+;)*\w+:\d+\):\d+$`)
	var a Production
	body, err := ioutil.ReadFile("../" + file)
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
	fmt.Println("No more process doable at cycle : " + strconv.Itoa(prod.current_cycle))
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
	fmt.Println(*Chaine)
	if len(Chaine.processes) == 0 {
		fmt.Println("Missing processes")
		os.Exit(0)
	}
	for {
		Chaine.resolve()
	}
}
