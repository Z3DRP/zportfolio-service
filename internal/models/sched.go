package models

import (
	"fmt"
	"slices"
	"sort"
	"sync"
	"time"
)

const LOWER_HOUR_BOUND = 8
const UPPER_HOUR_BOUND = 17
const LOWER_MIN_BOUND = 0
const UPPER_MIN_BOUND = 0
const LOWER_WEEK_BOUND = 0
const UPPER_WEEK_BOUND = 4
const STD_DAY = 1
const STD_MONTH = 9
const STD_SEC = 0

//    mon tue wen thur fri
// 8  t1  x   t2   t3  x
// 9  t4  t5  x    x   x

type HourlySchedule struct {
	Hour       int
	Hour12fmt  string
	TasksByDay map[int]Tasklist
}

func newHourlySchedule(hr int, fmtHr string, tasks map[int]Tasklist) *HourlySchedule {
	return &HourlySchedule{
		Hour:       hr,
		Hour12fmt:  fmtHr,
		TasksByDay: tasks,
	}
}

func newHourlyAgenda(avbs []Availability, tasks Tasklist) []HourlySchedule {
	sortedDays := sortAvailability(avbs)
	hours := make(map[int]HourlySchedule)
	hrlySchedule := make([]HourlySchedule, 0)
	for h := LOWER_HOUR_BOUND; h <= UPPER_HOUR_BOUND; h++ {
		if _, ok := hours[h]; !ok {
			hours[h] = HourlySchedule{}
		}
	}

	for _, avb := range avbs {
		if _, ok := hours[avb.HourKey()]; !ok {
			hours[avb.HourKey()] = *newHourlySchedule(avb.HourKey(), Get12HrFmt(avb.AvailableFrom), make(map[int]Tasklist, 0))
		}
	}

	for hr, schedule := range hours {
		schedule.TasksByDay = createDailyHrTasks(tasks, hr, sortedDays)
	}

	for _, agenda := range hours {
		hrlySchedule = append(hrlySchedule, agenda)
	}

	sortHourlySchedule(hrlySchedule)
	return hrlySchedule
}

func createDailyHrTasks(tasks Tasklist, hr int, avbDays []int) map[int]Tasklist {
	mday := Monday
	fday := Friday
	days := make(map[int]Tasklist, 0)

	for day := mday; day <= fday; day++ {
		if _, ok := days[day]; !ok {
			days[day] = make(Tasklist, 0)
		}
	}

	// NOTE days is mon-fri 1-5 indx to a list of tasks for that day of a specific time
	for _, task := range tasks {
		if slices.Contains(avbDays, task.StartTime.Day()) && slices.Contains(avbDays, task.EndTime.Day()) {
			// TODO grab all tasks for specified time put them in their day bucket then sort bucket and then return them
			if tklist, ok := days[task.StartTime.Day()]; ok {
				if task.StartTime.Hour() == hr {
					tklist = append(tklist, task)
				}
			}
		}
	}

	for _, tlist := range days {
		if len(tlist) > 0 {
			sortTasks(tlist)
		}
	}
	return days
}

type Schedule struct {
	Availability   []Availability
	HourlyAgenda   []HourlySchedule
	DaysAvailable  map[int]bool
	HoursAvailable map[string]bool
}

func NewSchedule(avb []Availability, tasks Tasklist) *Schedule {
	return &Schedule{
		Availability:   avb,
		HourlyAgenda:   newHourlyAgenda(avb, tasks),
		DaysAvailable:  calcDailyAvailability(avb),
		HoursAvailable: calcHourlyAvailability(avb),
	}
}

func calcHourlyAvailability(daysAvailabile []Availability) map[string]bool {
	// if from.Hour() < LOWER_HOUR_BOUND {
	// 	OutOfBounds := errors.New("availabile from time out of bounds")
	// 	return nil, OutOfBounds
	// }
	// if to.Hour() > UPPER_HOUR_BOUND {
	// 	OutOfBounds := errors.New("availabile to time out of bounds")
	// 	return nil, OutOfBounds
	// }
	// if from.Hour() > to.Hour() {
	// 	InvalidTimes := errors.New("from time must be before to time")
	// 	return nil, InvalidTimes
	// }
	hrRange := make(map[string]bool, 0)
	for _, day := range daysAvailabile {
		if day.AvailableFrom.Hour() >= LOWER_HOUR_BOUND && day.AvailableTo.Hour() <= UPPER_HOUR_BOUND {
			if _, exists := hrRange[MakeCompositeKey(day.WeekDayKey(), day.HourKey())]; !exists {
				hrRange[MakeCompositeKey(day.WeekDayKey(), day.HourKey())] = true
			}
		}
	}
	return hrRange
}

func calcDailyAvailability(daysAvailabile []Availability) map[int]bool {
	// TODO when making availabiltyRange for daily loop over dailyAvaialbilities here to getAll
	tmRange := make(map[int]bool, UPPER_HOUR_BOUND-LOWER_HOUR_BOUND)
	for i := LOWER_WEEK_BOUND; i <= UPPER_WEEK_BOUND; i++ {
		if _, exists := tmRange[i]; !exists {
			tmRange[i] = false
		}
	}

	for _, day := range daysAvailabile {
		if _, exists := tmRange[day.WeekDayKey()]; exists {
			tmRange[day.WeekDayKey()] = true
		}
	}

	return tmRange
}

func sortAvailability(avbs []Availability) []int {
	days := make([]int, 0)
	for _, avb := range avbs {
		days = append(days, avb.Day)
	}

	alen := len(days)
	for indx := alen/2 - 1; indx < 0; indx-- {
		heapifyDays(days, alen, indx)
	}

	for indx := alen - 1; indx >= 0; indx-- {
		days[indx], days[0] = days[0], days[indx]
		heapifyDays(days, indx, 0)
	}
	return days
}

func heapifyDays(days []int, length, indx int) {
	largest := indx
	leftChild := 2*indx + 1
	rightChild := 2*indx + 2

	if leftChild < length && days[indx] < days[leftChild] {
		largest = leftChild
	}

	if rightChild < length && days[indx] < days[rightChild] {
		largest = rightChild
	}

	if largest != indx {
		days[indx], days[largest] = days[largest], days[indx]
		heapifyDays(days, length, largest)
	}
}

func sortHourlySchedule(sch []HourlySchedule) {
	slen := len(sch)
	for indx := slen/2 - 1; indx < 0; indx-- {
		heapifySchedule(sch, slen, indx)
	}

	for indx := slen - 1; indx >= 0; indx-- {
		sch[indx], sch[0] = sch[0], sch[indx]
		heapifySchedule(sch, indx, 0)
	}
}

func heapifySchedule(sc []HourlySchedule, length, indx int) {
	largest := indx
	leftChild := 2*indx + 1
	rightChild := 2*indx + 2

	if leftChild < largest && sc[indx].Hour < sc[leftChild].Hour {
		largest = leftChild
	}

	if rightChild < largest && sc[indx].Hour < sc[rightChild].Hour {
		largest = rightChild
	}

	if largest != indx {
		sc[indx], sc[largest] = sc[largest], sc[indx]
		heapifySchedule(sc, length, largest)
	}
}

func sortTasks(tasks Tasklist) {
	tlen := len(tasks)

	for indx := tlen/2 - 1; indx < 0; indx-- {
		heapifyTasks(tasks, tlen, indx)
	}

	for indx := tlen - 1; tlen >= 0; indx-- {
		tasks[indx], tasks[0] = tasks[0], tasks[indx]
		heapifyTasks(tasks, tlen, indx)
	}
}

func heapifyTasks(tlist Tasklist, length, indx int) {
	largest := indx
	leftChild := indx*2 + 1
	rightChild := indx*2 + 2

	if leftChild < length && tlist[indx].StartTime.Before(tlist[leftChild].StartTime) {
		largest = leftChild
	}

	if rightChild < length && tlist[indx].StartTime.Before(tlist[rightChild].StartTime) {
		largest = rightChild
	}

	if largest != indx {
		tlist[indx], tlist[largest] = tlist[largest], tlist[indx]
		heapifyTasks(tlist, length, largest)
	}
}

func getAntemeridiem(hr int) string {
	antemeridiem := "AM"
	if hr >= 12 {
		antemeridiem = "PM"
	}
	return antemeridiem
}

func convertHr(h int) int {
	hr := h % 12
	if hr == 0 {
		hr = 12
	}
	return hr
}

func GetKey12HrFmt(key int) string {
	hr := convertHr(key)
	antemeridiem := getAntemeridiem(hr)
	return fmt.Sprintf("%02d %s", hr, antemeridiem)
}

func Get12HrFmt(t time.Time) string {
	hr := convertHr(t.Hour())
	antem := getAntemeridiem(hr)
	return fmt.Sprintf("%v:%v %s", hr, t.Minute(), antem)
}

func FmtTimeSpan(start, end time.Time) string {
	stm := Get12HrFmt(start)
	etm := Get12HrFmt(end)
	return fmt.Sprintf("%s - %s", stm, etm)
}

func SortTaskByHour(tasks Tasklist) map[int]Tasklist {
	tasksByHr := make(map[int]Tasklist, UPPER_HOUR_BOUND-LOWER_HOUR_BOUND)
	keys := make([]int, 0, len(tasksByHr))

	for hr := LOWER_HOUR_BOUND; hr <= UPPER_HOUR_BOUND; hr++ {
		if _, exists := tasksByHr[hr]; !exists {
			tasksByHr[hr] = make(Tasklist, 0)
			keys = append(keys, hr)
		}
	}

	if len(tasks) == 0 {
		return tasksByHr
	}

	for _, task := range tasks {
		if isTimeInRange(task.StartTime, task.EndTime) {
			tasksByHr[task.WeekDay()] = append(tasksByHr[task.WeekDay()], task)
		}
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, k := range keys {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			sort.Slice(tasksByHr[k], func(curIndx, nxtIndx int) bool {
				return tasksByHr[k][curIndx].StartTime.Compare(tasksByHr[k][nxtIndx].StartTime) == -1
			})
			mu.Unlock()
		}()
	}
	wg.Wait()
	return tasksByHr
}

func SortTasksByDay(tasks Tasklist) map[int]Tasklist {
	tasksByDay := make(map[int]Tasklist, 7)
	keys := make([]int, 0, len(tasksByDay))
	for day := LOWER_WEEK_BOUND; day <= UPPER_WEEK_BOUND; day++ {
		if _, exists := tasksByDay[day]; !exists {
			tasksByDay[day] = make(Tasklist, 0)
			keys = append(keys, day)
		}
	}

	if len(tasks) == 0 {
		return tasksByDay
	}

	for _, task := range tasks {
		if isDayInRange(task.StartTime, task.EndTime) {
			tasksByDay[task.WeekDay()] = append(tasksByDay[task.WeekDay()], task)
		}
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, key := range keys {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			sort.Slice(tasksByDay[key], func(curIndx, nxtIndx int) bool {
				return tasksByDay[key][curIndx].StartTime.Compare(tasksByDay[key][nxtIndx].StartTime) == -1
			})
			mu.Unlock()
		}()
	}
	wg.Wait()

	return tasksByDay
}

func GenTempAvail() []Availabler {
	return []Availabler{
		NewAvailableDay(1, 9, 0, 5, 0),
		NewAvailableDay(2, 9, 0, 5, 0),
		NewAvailableDay(3, 9, 0, 5, 0),
		NewAvailableDay(4, 9, 0, 5, 0),
		NewAvailableDay(5, 9, 0, 5, 0),
	}
}

func NewTempAvailability(daysOpen []AvailableDay) []Availabler {
	openDays := make([]Availabler, 0)
	for _, day := range daysOpen {
		openDays = append(openDays, NewAvailability(day))
	}
	return openDays
}

func MakeCompositeKey(k1, k2 int) string {
	return fmt.Sprintf("%v%v", k1, k2)
}

func standardizeDateTime(hr, min, sec, nansec int) time.Time {
	return time.Date(time.Now().Year(), STD_MONTH, STD_DAY, hr, min, sec, nansec, time.UTC)
}

func standardizeDate(day int) time.Time {
	return time.Date(time.Now().Year(), STD_MONTH, day, LOWER_HOUR_BOUND, LOWER_MIN_BOUND, STD_SEC, STD_SEC, time.UTC)
}

func isTimeInRange(start, end time.Time) bool {
	isInRange := false
	startTime := standardizeDateTime(start.Hour(), start.Minute(), start.Second(), start.Nanosecond())
	endTime := standardizeDateTime(end.Hour(), end.Minute(), end.Second(), end.Nanosecond())
	lwBoundTime := standardizeDateTime(LOWER_HOUR_BOUND, LOWER_MIN_BOUND, STD_SEC, STD_SEC)
	upBoundTime := standardizeDateTime(UPPER_HOUR_BOUND, UPPER_MIN_BOUND, STD_SEC, STD_SEC)

	if startTime.Compare(lwBoundTime) >= 0 && startTime.Compare(upBoundTime) <= 0 {
		isInRange = true
	}

	if endTime.Compare(startTime) == 0 && endTime.Compare(lwBoundTime) >= 0 && endTime.Compare(upBoundTime) <= 0 {
		isInRange = true
	}

	return isInRange
}

func isDayInRange(start, end time.Time) bool {
	isInRange := false
	startDate := standardizeDate(start.Day())
	endDate := standardizeDate(end.Day())
	lwBoundDate := standardizeDate(LOWER_WEEK_BOUND)
	upBoundDate := standardizeDate(UPPER_WEEK_BOUND)

	if startDate.Compare(lwBoundDate) >= 0 && startDate.Compare(upBoundDate) < 0 {
		isInRange = true
	}

	if endDate.Compare(startDate) == 0 && endDate.Compare(lwBoundDate) >= 0 && endDate.Compare(upBoundDate) <= 0 {
		isInRange = true
	}

	return isInRange
}
