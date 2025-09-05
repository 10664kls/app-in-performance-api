package appin

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Overview is the overview of App-In.
type Overview struct {
	// ActiveExecutor is the number of active executor.
	ActiveExecutor int64 `json:"activeExecutor"`

	// TopPerformer is the top performer who performed the most App-In.
	TopPerformer *TopPerformer `json:"topPerformer"`

	// Conversion is the conversion rate.
	Conversion *Conversion `json:"conversion"`

	// TimeIntervalsByConverted is the time intervals for converted App-In.
	TimeIntervalsByConverted []*TimeInterval `json:"timeIntervalsByConverted"`

	// TimeIntervalsByPending is the time intervals for pending App-In.
	TimeIntervalsByPending []*TimeInterval `json:"timeIntervalsByPending"`

	// BestTimeUsed is the executor with the best time used for App-In.
	BestTimeUsed *BestTimeExecutor `json:"bestTimeUsed"`

	// Leaderboard is the leaderboard of App-In.
	// Top 5 performers
	Leaderboards []*Leaderboard `json:"leaderboards"`

	// ProductMetrics is the product metrics of App-In.
	ProductMetrics []*ProductMetrics `json:"productMetrics"`

	// CAFinalOverview is the CA operation performed by App-In.
	CAFinalOverview *CAFinalOverview `json:"caFinalOverview"`
}

func newOverview(appIns []*AppIn) *Overview {
	groups := groupAppInByExecutor(appIns)
	performances := calculatePerformanceConversionMetricsByExecutor(groups)
	o := new(Overview)

	o.ActiveExecutor = int64(len(groups))
	o.TopPerformer = getTopPerformer(performances)
	o.Conversion = newConversion(appIns)
	o.Leaderboards = createLeaderboards(performances)

	o.TimeIntervalsByConverted = createTimeIntervalsByConverted(appIns)
	o.BestTimeUsed = findBestTimeUsedByExecutor(performances)

	o.TimeIntervalsByPending = createTimeIntervalsByPending(appIns)
	o.ProductMetrics = calculatePerformanceConversionMetricsByProduct(appIns)

	return o
}

// SetCAFinal sets the CA operation performed by App-In.
func (o *Overview) SetCAFinal(ca []*CAFinal) {
	o.CAFinalOverview = newCAFinalOverview(ca)
}

func newCAFinalOverview(appins []*CAFinal) *CAFinalOverview {
	groups := groupCAFinalByExecutor(appins)
	performances := calculateCAFinalConversionMetricsByExecutor(groups)

	c := new(CAFinalOverview)
	c.ActiveExecutor = int64(len(groups))
	c.TopPerformer = getTopPerformer(performances)
	c.Leaderboards = createLeaderboards(performances)
	c.BestTimeUsed = findBestTimeUsedByExecutor(performances)

	c.Conversion = newCAFinalConversion(appins)
	c.TimeIntervalsByConverted = createCAFinalTimeIntervalsByConverted(appins)
	c.TimeIntervalsByPending = createCAFinalTimeIntervalsByPending(appins)

	return c
}

// CAFinalOverview is the CA operation performed by App-In.
type CAFinalOverview struct {
	// ActiveExecutor is the number of active executor.
	ActiveExecutor int64 `json:"activeExecutor"`

	// TopPerformer is the top performer who performed the most App-In.
	TopPerformer *TopPerformer `json:"topPerformer"`

	// Conversion is the conversion rate.
	Conversion *Conversion `json:"conversion"`

	// TimeIntervalsByConverted is the time intervals for converted App-In.
	TimeIntervalsByConverted []*TimeInterval `json:"timeIntervalsByConverted"`

	// TimeIntervalsByPending is the time intervals for pending App-In.
	TimeIntervalsByPending []*TimeInterval `json:"timeIntervalsByPending"`

	// BestTimeUsed is the executor with the best time used for App-In.
	BestTimeUsed *BestTimeExecutor `json:"bestTimeUsed"`

	// Leaderboard is the leaderboard of App-In.
	// Top 5 performers
	Leaderboards []*Leaderboard `json:"leaderboards"`
}

// BestTimeExecutor is the executor with the best time used for App-In.
type BestTimeExecutor struct {
	// DisplayName is the display name of the executor.
	DisplayName string `json:"displayName"`

	// BestTime is the best time used for App-In.
	BestTime time.Duration `json:"bestTime"`
}

// Conversion is the conversion rate.
type Conversion struct {
	// Total is the total number of App-In.
	Total int64 `json:"total"`

	// Converted is the number of App-In converted.
	Converted int64 `json:"converted"`

	// NotPassed is the number of App-In not passed.
	NotPassed int64 `json:"notPassed"`

	// Rate is the conversion rate.
	Rate float32 `json:"rate"`

	// Fastest is the number of App-In performed in the shortest time. Under 30min.
	Fastest int64 `json:"fastest"`

	// FastestPercent is the percentage of App-In performed in the shortest time.
	FastestPercent float32 `json:"fastestPercent"`

	// Slowest is the number of App-In performed in the longest time. Over 5h.
	NeedAttention int64 `json:"needAttention"`

	// BestTime is the best time used for App-In.
	BestTime time.Duration `json:"bestTime"`

	// AverageTime is the average time for App-in performed.
	AverageTime time.Duration `json:"averageTime"`
}

// TimeInterval is the time interval for App-In.
type TimeInterval struct {
	// Title is the title of the time interval. ex: "1h", "1d", "1w"
	Title string `json:"title"`

	// Total is the total number of App-In.
	Total int64 `json:"total"`
}

type Leaderboard struct {
	// Rank is the rank of the performer.
	Rank int64 `json:"rank"`

	// DisplayName is the display name of the performer.
	DisplayName string `json:"displayName"`

	// Total is the total number of App-In that the performer handled.
	Total int64 `json:"total"`

	// Converted is the number of App-In converted by the performer.
	Converted int64 `json:"converted"`

	// ConversionRate is the conversion rate of the performer.
	ConversionRate float32 `json:"conversionRate"`

	// AverageTime is the average time for App-in performed by the performer.
	AverageTime time.Duration `json:"averageTime"`

	// BestTime is the best time used for App-In by the performer.
	BestTime time.Duration `json:"bestTime"`

	// Performances is the performance of the performer for each time interval.
	Performances []*TimeInterval `json:"performances"`
}

// ProductMetrics is the product metrics.
type ProductMetrics struct {
	// Name is the name of the product.
	Name string `json:"name"`

	// Total is the total number of App-In for the product.
	Total int64 `json:"total"`

	// Converted is the number of App-In converted for the product.
	Converted int64 `json:"converted"`

	// NotPassed is the number of App-In not passed for the product.
	NotPassed int64 `json:"notPassed"`

	// ConversionRate is the conversion rate for the product.
	ConversionRate float32 `json:"conversionRate"`

	// AverageTime is the average time for App-in performed for the product.
	AverageTime time.Duration `json:"averageTime"`
}

// TopPerformer is the top performer.
type TopPerformer struct {
	// DisplayName is the display name of the top performer.
	DisplayName string `json:"displayName"`

	// Converted is the number of App-In converted by the top performer.
	Converted int64 `json:"converted"`

	// ConversionRate is the conversion rate of the top performer.
	ConversionRate float32 `json:"conversionRate"`
}

// newConversion calculates and returns conversion metrics from a slice of AppIn.
func newConversion(appIns []*AppIn) *Conversion {
	total := int64(len(appIns))

	var sum, bestTime time.Duration
	var fastestCount, needAttention, converted, notPassed int64
	const threshold = 5 * time.Hour

	for _, appIn := range appIns {
		status := strings.ToLower(appIn.Status)
		if len(status) > 0 && !strings.Contains(status, "not pass") && appIn.CompletedAt != nil {
			converted++
			duration := appIn.CompletedAt.Sub(appIn.CreatedAt)
			sum += duration

			if duration <= time.Minute*30 {
				fastestCount++
			}

			if (bestTime == 0 || duration < bestTime) && duration >= time.Minute {
				bestTime = duration
			}
		}

		if status == "" && appIn.CompletedAt == nil {
			duration := time.Since(appIn.CreatedAt)
			if duration > threshold {
				needAttention++
			}
		}

		if strings.Contains(status, "not pass") {
			notPassed++
		}
	}

	var conversionRate, fastestPercent float32
	var averageTime time.Duration

	processed := converted + notPassed
	if total > 0 {
		conversionRate = float32(processed) / float32(total) * 100
		if processed > 0 {
			averageTime = sum / time.Duration(processed)
		}
	}

	if processed > 0 {
		fastestPercent = float32(fastestCount) / float32(processed) * 100
	}

	return &Conversion{
		Total:          total,
		Converted:      converted,
		Rate:           conversionRate,
		AverageTime:    averageTime,
		NeedAttention:  needAttention,
		Fastest:        fastestCount,
		BestTime:       bestTime,
		FastestPercent: fastestPercent,
		NotPassed:      notPassed,
	}
}

func newCAFinalConversion(cs []*CAFinal) *Conversion {
	total := int64(len(cs))

	var sum, bestTime time.Duration
	var fastestCount, needAttention, converted int64
	const threshold = 5 * time.Hour

	for _, c := range cs {
		status := strings.ToLower(c.Status)
		if len(status) > 0 && status == "completed" && c.CompletedAt != nil {
			converted++
			duration := c.CompletedAt.Sub(c.CreatedAt)
			sum += duration

			if duration <= time.Minute*30 {
				fastestCount++
			}

			if (bestTime == 0 || duration < bestTime) && duration >= time.Minute {
				bestTime = duration
			}
		}

		if status != "completed" && c.CompletedAt == nil {
			duration := time.Since(c.CreatedAt)
			if duration > threshold {
				needAttention++
			}
		}
	}

	var conversionRate, fastestPercent float32
	var averageTime time.Duration

	if total > 0 {
		conversionRate = float32(converted) / float32(total) * 100
		if converted > 0 {
			averageTime = sum / time.Duration(converted)
		}
	}

	if converted > 0 {
		fastestPercent = float32(fastestCount) / float32(converted) * 100
	}

	return &Conversion{
		Total:          total,
		Converted:      converted,
		Rate:           conversionRate,
		AverageTime:    averageTime,
		NeedAttention:  needAttention,
		Fastest:        fastestCount,
		BestTime:       bestTime,
		FastestPercent: fastestPercent,
	}
}

// getTopPerformer returns the top performer from a slice of AppIn.
func getTopPerformer(conversions map[string]*performerMetric) *TopPerformer {
	var top TopPerformer

	for executor, c := range conversions {
		if c.Conversion.Converted > top.Converted || (c.Conversion.Converted == top.Converted && c.Conversion.Rate > top.ConversionRate) {
			top.DisplayName = executor
			top.Converted = c.Conversion.Converted
			top.ConversionRate = c.Conversion.Rate
		}
	}

	return &top
}

func groupAppInByExecutor(appIns []*AppIn) map[string][]*AppIn {
	groups := make(map[string][]*AppIn, 0)
	for _, a := range appIns {
		if a.Executor == "" {
			continue
		}

		groups[a.Executor] = append(groups[a.Executor], a)
	}

	return groups
}

func groupCAFinalByExecutor(appIns []*CAFinal) map[string][]*CAFinal {
	groups := make(map[string][]*CAFinal, 0)
	for _, a := range appIns {
		if a.Executor == "" {
			continue
		}

		groups[a.Executor] = append(groups[a.Executor], a)
	}

	return groups
}

func groupAppInByProduct(appIns []*AppIn) map[string][]*AppIn {
	groups := make(map[string][]*AppIn, 0)
	for _, a := range appIns {
		if a.Type == "" {
			continue
		}

		switch strings.ToLower(a.Product) {
		case "sale auto":
			switch {
			case strings.Contains(strings.ToLower(a.Type), "c4c"):
				key := fmt.Sprintf(`C4C | %s`, a.Product)
				groups[key] = append(groups[key], a)

			case strings.Contains(strings.ToLower(a.Type), "used car"):
				key := fmt.Sprintf(`Used Car | %s`, a.Product)
				groups[key] = append(groups[key], a)

			case strings.Contains(strings.ToLower(a.Type), "mc"):
				key := fmt.Sprintf(`MC | %s`, a.Product)
				groups[key] = append(groups[key], a)

			default:
				groups[a.Product] = append(groups[a.Product], a)
			}

		default:
			groups[a.Product] = append(groups[a.Product], a)
		}

	}

	return groups
}

func calculatePerformanceConversionMetricsByProduct(appIns []*AppIn) []*ProductMetrics {
	groups := groupAppInByProduct(appIns)
	products := make([]*ProductMetrics, 0)

	for product, apps := range groups {
		c := newConversion(apps)
		products = append(products, &ProductMetrics{
			Name:           product,
			Total:          int64(len(apps)),
			Converted:      c.Converted,
			ConversionRate: c.Rate,
			AverageTime:    c.AverageTime,
			NotPassed:      c.NotPassed,
		})
	}

	sort.Slice(products, func(i, j int) bool {
		// Sort by Converted count (descending)
		if products[i].Converted != products[j].Converted {
			return products[i].Converted > products[j].Converted
		}

		// If Converted count is the same, sort by Conversion rate (descending)
		if products[i].ConversionRate != products[j].ConversionRate {
			return products[i].ConversionRate > products[j].ConversionRate
		}

		// If rates are also the same, sort by average time (ascending)
		return products[i].AverageTime < products[j].AverageTime
	})

	return products
}

func calculatePerformanceConversionMetricsByExecutor(groups map[string][]*AppIn) map[string]*performerMetric {
	performers := make(map[string]*performerMetric, 0)

	for executor, apps := range groups {
		performers[executor] = &performerMetric{
			DisplayName:  executor,
			Conversion:   newConversion(apps),
			Performances: createTimeIntervalsByConverted(apps),
		}
	}

	return performers
}

func calculateCAFinalConversionMetricsByExecutor(groups map[string][]*CAFinal) map[string]*performerMetric {
	performers := make(map[string]*performerMetric, 0)

	for executor, cs := range groups {
		performers[executor] = &performerMetric{
			DisplayName:  executor,
			Conversion:   newCAFinalConversion(cs),
			Performances: createCAFinalTimeIntervalsByConverted(cs),
		}
	}

	return performers
}

type performerMetric struct {
	DisplayName  string
	Conversion   *Conversion
	Performances []*TimeInterval
}

// createLeaderboards creates a leaderboard from the conversion metrics. And sorts it by conversion number and rate.
// Will return the top 5 performers.
func createLeaderboards(conversions map[string]*performerMetric) []*Leaderboard {
	performers := make([]*performerMetric, 0, len(conversions))
	for _, c := range conversions {
		performers = append(performers, c)
	}

	sort.Slice(performers, func(i, j int) bool {
		// Sort by Converted count (descending)
		if performers[i].Conversion.Converted != performers[j].Conversion.Converted {
			return performers[i].Conversion.Converted > performers[j].Conversion.Converted
		}

		// If Converted count is the same, sort by Conversion rate (descending)
		if performers[i].Conversion.Rate != performers[j].Conversion.Rate {
			return performers[i].Conversion.Rate > performers[j].Conversion.Rate
		}

		// If rates are also the same, sort by average time (ascending)
		return performers[i].Conversion.AverageTime < performers[j].Conversion.AverageTime
	})

	leaderboards := make([]*Leaderboard, 0, 5)
	for i, p := range performers {
		leaderboards = append(leaderboards, &Leaderboard{
			Rank:           int64(i + 1),
			DisplayName:    p.DisplayName,
			Converted:      p.Conversion.Converted,
			Total:          p.Conversion.Total,
			ConversionRate: p.Conversion.Rate,
			Performances:   p.Performances,
			AverageTime:    p.Conversion.AverageTime,
			BestTime:       p.Conversion.BestTime,
		})

		// Limit to top 5
		if len(leaderboards) >= 5 {
			break
		}
	}

	return leaderboards
}

func createTimeIntervalsByConverted(appIns []*AppIn) []*TimeInterval {
	titles := []string{
		"<30min",
		"<1h",
		"<2h",
		"<3h",
		"<4h",
		"<5h",
		"5h+",
	}

	intervals := make(map[string]int64)
	for _, t := range titles {
		intervals[t] = 0
	}

	for _, a := range appIns {
		status := strings.ToLower(a.Status)
		if len(a.Status) > 0 && !strings.Contains(status, "not pass") && a.CompletedAt != nil {
			duration := a.CompletedAt.Sub(a.CreatedAt)
			switch {
			case duration < 30*time.Minute:
				intervals["<30min"]++
			case duration < time.Hour:
				intervals["<1h"]++
			case duration < 2*time.Hour:
				intervals["<2h"]++
			case duration < 3*time.Hour:
				intervals["<3h"]++
			case duration < 4*time.Hour:
				intervals["<4h"]++
			case duration < 5*time.Hour:
				intervals["<5h"]++
			default:
				intervals["5h+"]++
			}
		}
	}

	ts := make([]*TimeInterval, 0)
	for _, t := range titles {
		ts = append(ts, &TimeInterval{
			Title: t,
			Total: intervals[t],
		})
	}

	return ts
}

func createTimeIntervalsByPending(appIns []*AppIn) []*TimeInterval {
	titles := []string{
		"<30min",
		"<1h",
		"<2h",
		"<3h",
		"<4h",
		"<5h",
		"5h+",
	}

	intervals := make(map[string]int64)
	for _, t := range titles {
		intervals[t] = 0
	}

	now := time.Now()
	for _, a := range appIns {
		if a.Status == "" && a.CompletedAt == nil {
			duration := now.Sub(a.CreatedAt)
			switch {
			case duration < 30*time.Minute:
				intervals["<30min"]++
			case duration < time.Hour:
				intervals["<1h"]++
			case duration < 2*time.Hour:
				intervals["<2h"]++
			case duration < 3*time.Hour:
				intervals["<3h"]++
			case duration < 4*time.Hour:
				intervals["<4h"]++
			case duration < 5*time.Hour:
				intervals["<5h"]++
			default:
				intervals["5h+"]++
			}
		}
	}

	ts := make([]*TimeInterval, 0)
	for _, t := range titles {
		ts = append(ts, &TimeInterval{
			Title: t,
			Total: intervals[t],
		})
	}

	return ts
}

func findBestTimeUsedByExecutor(p map[string]*performerMetric) *BestTimeExecutor {
	if len(p) == 0 {
		return &BestTimeExecutor{
			DisplayName: "N/A",
			BestTime:    0,
		}
	}

	var bestExecutor string
	var bestTime time.Duration

	for executor, metric := range p {
		if metric.Conversion.BestTime > 0 {
			if bestTime == 0 || metric.Conversion.BestTime < bestTime {
				bestTime = metric.Conversion.BestTime
				bestExecutor = executor
			}
		}
	}

	return &BestTimeExecutor{
		DisplayName: bestExecutor,
		BestTime:    bestTime,
	}
}

func createCAFinalTimeIntervalsByConverted(cs []*CAFinal) []*TimeInterval {
	titles := []string{
		"<30min",
		"<1h",
		"<2h",
		"<3h",
		"<4h",
		"<5h",
		"5h+",
	}

	intervals := make(map[string]int64)
	for _, t := range titles {
		intervals[t] = 0
	}

	for _, a := range cs {
		status := strings.ToLower(a.Status)
		if status == "completed" && a.CompletedAt != nil {
			duration := a.CompletedAt.Sub(a.CreatedAt)
			switch {
			case duration < 30*time.Minute:
				intervals["<30min"]++
			case duration < time.Hour:
				intervals["<1h"]++
			case duration < 2*time.Hour:
				intervals["<2h"]++
			case duration < 3*time.Hour:
				intervals["<3h"]++
			case duration < 4*time.Hour:
				intervals["<4h"]++
			case duration < 5*time.Hour:
				intervals["<5h"]++
			default:
				intervals["5h+"]++
			}
		}
	}

	ts := make([]*TimeInterval, 0)
	for _, t := range titles {
		ts = append(ts, &TimeInterval{
			Title: t,
			Total: intervals[t],
		})
	}

	return ts
}

func createCAFinalTimeIntervalsByPending(cs []*CAFinal) []*TimeInterval {
	titles := []string{
		"<30min",
		"<1h",
		"<2h",
		"<3h",
		"<4h",
		"<5h",
		"5h+",
	}

	intervals := make(map[string]int64)
	for _, t := range titles {
		intervals[t] = 0
	}

	now := time.Now()
	for _, a := range cs {
		if strings.ToLower(a.Status) != "completed" && a.CompletedAt == nil {
			duration := now.Sub(a.CreatedAt)
			switch {
			case duration < 30*time.Minute:
				intervals["<30min"]++
			case duration < time.Hour:
				intervals["<1h"]++
			case duration < 2*time.Hour:
				intervals["<2h"]++
			case duration < 3*time.Hour:
				intervals["<3h"]++
			case duration < 4*time.Hour:
				intervals["<4h"]++
			case duration < 5*time.Hour:
				intervals["<5h"]++
			default:
				intervals["5h+"]++
			}
		}
	}

	ts := make([]*TimeInterval, 0)
	for _, t := range titles {
		ts = append(ts, &TimeInterval{
			Title: t,
			Total: intervals[t],
		})
	}

	return ts
}
