package storage

import (
	"encoding/json"
	"fmt"
	homedir "github.com/mitchellh/go-homedir"
	"io/ioutil"
	"math"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"text/template"
	"time"
)

var shortTasksTemplate string = `
Id  TimeSpent Title
{{range .Tasks}}{{.Id | printf "%-3d"}} {{.Events | timeSpent}}  {{.Title}}
{{end}}
`

var verboseTasksTemplate string = `{{range .Tasks}}------------------------------------------------
[{{.Id}}] {{.Title}}

{{.Description}}

Time Created: {{.CreateTime | fmtTime}}
Time Started: {{.StartTime | fmtTime}}
Time Ended: {{.EndTime | fmtTime}}
Time Spent: {{.Events | timeSpent}}

Elapsed  Start                End
{{range .Events}}{{. | timeElapsed}} {{.StartTime | fmtTime}} {{.EndTime | fmtTime}}
{{end}}
{{end}}`

type Event struct {
	StartTime time.Time
	EndTime   time.Time
}

type Task struct {
	Id          int
	Title       string
	Description string
	Events      []Event
	CreateTime  time.Time
	StartTime   time.Time
	EndTime     time.Time
}

type Tasks struct {
	Tasks []Task
}

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

func setDefaultTaskValues(task *Task) {
	task.Events = []Event{}
	task.CreateTime = time.Now()
}

// TODO: Should refactor entire package and separate data layer from modeling layer
func AddTask(task Task) int {
	setDefaultTaskValues(&task)

	tasks := loadTasks()

	id := 1

	for _, t := range tasks.Tasks {
		if t.Id >= id {
			id = t.Id + 1
		}
	}
	task.Id = id

	tasks.Tasks = append(tasks.Tasks, task)

	saveTasks(tasks)

	return id
}

func DeleteTask(id int) {
	tasks := loadTasks()

	for i, t := range tasks.Tasks {
		if t.Id == id {
			tasks.Tasks = remove(tasks.Tasks, i)
			break
		}
	}

	saveTasks(tasks)
}

func StartTask(id int) {
	tasks := loadTasks()

	var task *Task

	var index int
	for i, t := range tasks.Tasks {
		if t.Id == id {
			index = i
			task = &t
			break
		}
	}

	if task == nil {
		panic("Task " + string(id) + " does not exist.")
	}

	startTime := time.Now()

	if isTimeUnset(task.StartTime) {
		task.StartTime = startTime
	}

	// TODO: Skip creating event if already started
	event := Event{
		StartTime: startTime,
	}
	task.Events = append(task.Events, event)
	tasks.Tasks[index] = *task

	saveTasks(tasks)

	isRunning := true
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		isRunning = false
		StopTask(id)
		fmt.Println("")
	}()

	for isRunning {
		printTimeElasped(startTime)
		time.Sleep(1 * time.Second)
	}
}

func StopTask(id int) {
	// TODO: Refactor shared code into function, i.e. finding a task, etc.
	tasks := loadTasks()

	var task *Task

	var index int
	for i, t := range tasks.Tasks {
		if t.Id == id {
			index = i
			task = &t
			break
		}
	}

	if task == nil {
		panic("Task " + string(id) + " does not exist.")
	}

	for i, e := range task.Events {
		if isTimeUnset(e.EndTime) {
			e.EndTime = time.Now()
			task.Events[i] = e
		}
	}
	tasks.Tasks[index] = *task

	saveTasks(tasks)
}

func ListTasks(verbose bool) {
	tasks := loadTasks()

	tmpl := shortTasksTemplate
	if verbose {
		tmpl = verboseTasksTemplate
	}
	s, err := template.New("tasks").
		Funcs(template.FuncMap{
			"fmtTime":     fmtTime,
			"timeElapsed": timeElapsed,
			"timeSpent":   timeSpent,
		}).
		Parse(tmpl)
	handleError(err)

	if err := s.Execute(os.Stdout, tasks); err != nil {
		handleError(err)
	}
}

func loadTasks() Tasks {
	tasks := Tasks{}
	raw_tasks, err := ioutil.ReadFile(getDBPath())
	if err != nil {
		if strings.HasSuffix(err.Error(), "no such file or directory") {
			tasks.Tasks = []Task{}
			return tasks
		}
		handleError(err)
	}

	err = json.Unmarshal([]byte(raw_tasks), &tasks)
	handleError(err)

	return tasks
}

func saveTasks(tasks Tasks) {
	s, err := json.MarshalIndent(tasks, "", "  ")
	handleError(err)

	err = ioutil.WriteFile(getDBPath(), s, 0644)
	handleError(err)
}

func isTimeUnset(t time.Time) bool {
	emptyTime := time.Time{}
	return emptyTime == t
}

func remove(slice []Task, i int) []Task {
	return append(slice[:i], slice[i+1:]...)
}

func getDBPath() string {
	home, err := homedir.Dir()
	handleError(err)

	return home + "/.clerk-db"
}

func printTimeElasped(startTime time.Time) {
	duration := time.Now().Sub(startTime)
	fmt.Printf("\rTime Elapsed: %s", toString(duration))
}

func fmtTime(t time.Time) string {
	if t == (time.Time{}) {
		return ""
	}

	return t.Format("Jan _2 15:04:05 2006")
}

func timeSpent(events []Event) string {
	var duration time.Duration = 0

	for _, e := range events {
		duration += getDuration(e)
	}

	return toString(duration)
}

func timeElapsed(e Event) string {
	return toString(getDuration(e))
}

func getDuration(e Event) time.Duration {
	if isTimeUnset(e.EndTime) {
		return time.Now().Sub(e.StartTime)
	}

	return e.EndTime.Sub(e.StartTime)
}

func toString(duration time.Duration) string {
	return fmt.Sprintf("%02.0f:%02.0f:%02.0f", getHours(duration), getMinutes(duration), getSeconds(duration))
}

func getHours(duration time.Duration) float64 {
	return math.Floor(duration.Hours())
}

func getMinutes(duration time.Duration) float64 {
	return math.Mod(duration.Minutes(), 60)
}

func getSeconds(duration time.Duration) float64 {
	return math.Mod(duration.Seconds(), 60)
}
