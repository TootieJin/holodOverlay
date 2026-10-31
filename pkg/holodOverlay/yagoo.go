package holodOverlay

import (
	"fmt"
	"io"
	"math"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/TootieJin/holodOverlay/pkg/sonolus"
)

type YagooFrame struct {
	Time  float64
	Score float64
}

type BpmChange struct {
	Beat float64
	Bpm  float64
}

var WEIGHT_MAP = map[string]float64{
	// Nanashi's Archetype
	"#BPM_CHANGE":    0,
	"Initialization": 0,
	"InputManager":   0,
	"Stage":          0,

	"NormalTapNote":   1,
	"CriticalTapNote": 1,

	"NormalFlickNote":   1.05,
	"CriticalFlickNote": 1.05,

	"NormalSlideStartNote":   1,
	"CriticalSlideStartNote": 1,

	"NormalSlideEndNote":   1,
	"CriticalSlideEndNote": 1,

	"NormalSlideEndFlickNote":   1.05,
	"CriticalSlideEndFlickNote": 1.05,

	"HiddenSlideTickNote":   0,
	"NormalSlideTickNote":   0.1,
	"CriticalSlideTickNote": 0.1,

	"IgnoredSlideTickNote":          0.1,
	"NormalAttachedSlideTickNote":   0.1,
	"CriticalAttachedSlideTickNote": 0.1,

	"NormalSlideConnector":   0,
	"CriticalSlideConnector": 0,

	"SimLine": 0,

	"NormalSlotEffect":       0,
	"SlideSlotEffect":        0,
	"FlickSlotEffect":        0,
	"CriticalSlotEffect":     0,
	"NormalSlotGlowEffect":   0,
	"SlideSlotGlowEffect":    0,
	"FlickSlotGlowEffect":    0,
	"CriticalSlotGlowEffect": 0,

	"NormalTraceNote":   1,
	"CriticalTraceNote": 1,

	"NormalTraceSlotEffect":     0,
	"NormalTraceSlotGlowEffect": 0,

	"DamageNote":           0.1,
	"DamageSlotEffect":     0,
	"DamageSlotGlowEffect": 0,

	"NormalTraceFlickNote":         1.05,
	"CriticalTraceFlickNote":       1.05,
	"NonDirectionalTraceFlickNote": 1.05,

	"NormalTraceSlideStartNote":   1,
	"NormalTraceSlideEndNote":     1,
	"CriticalTraceSlideStartNote": 1,
	"CriticalTraceSlideEndNote":   1,

	"TimeScaleGroup":  0,
	"TimeScaleChange": 0,

	// Next SEKAI's Archetype
	"#TIMESCALE_CHANGE": 0,
	"#TIMESCALE_GROUP":  0,
	"_InputManager":     0,
	"SlideManager":      0,
	"Connector":         0,
	"SlotGlowEffect":    0,
	"SlotEffect":        0,

	"NormalHeadTapNote":          1,
	"CriticalHeadTapNote":        1,
	"NormalHeadFlickNote":        1.05,
	"CriticalHeadFlickNote":      1.05,
	"NormalHeadTraceNote":        1,
	"CriticalHeadTraceNote":      1,
	"NormalHeadTraceFlickNote":   1.05,
	"CriticalHeadTraceFlickNote": 1.05,
	"NormalHeadReleaseNote":      1,
	"CriticalHeadReleaseNote":    1,

	"NormalTailTapNote":          1,
	"CriticalTailTapNote":        1,
	"NormalTailFlickNote":        1.05,
	"CriticalTailFlickNote":      1.05,
	"NormalTailTraceNote":        1,
	"CriticalTailTraceNote":      1,
	"NormalTailTraceFlickNote":   1.05,
	"CriticalTailTraceFlickNote": 1.05,
	"NormalTailReleaseNote":      1,
	"CriticalTailReleaseNote":    1,

	"TransientHiddenTickNote": 0.1,
	"NormalTickNote":          0.1,
	"CriticalTickNote":        0.1,
	"AnchorNote":              0,

	"FakeNormalTapNote":          0,
	"FakeCriticalTapNote":        0,
	"FakeNormalFlickNote":        0,
	"FakeCriticalFlickNote":      0,
	"FakeNormalTraceNote":        0,
	"FakeCriticalTraceNote":      0,
	"FakeNormalTraceFlickNote":   0,
	"FakeCriticalTraceFlickNote": 0,
	"FakeNormalReleaseNote":      0,
	"FakeCriticalReleaseNote":    0,

	"FakeNormalHeadTapNote":          0,
	"FakeCriticalHeadTapNote":        0,
	"FakeNormalHeadFlickNote":        0,
	"FakeCriticalHeadFlickNote":      0,
	"FakeNormalHeadTraceNote":        0,
	"FakeCriticalHeadTraceNote":      0,
	"FakeNormalHeadTraceFlickNote":   0,
	"FakeCriticalHeadTraceFlickNote": 0,
	"FakeNormalHeadReleaseNote":      0,
	"FakeCriticalHeadReleaseNote":    0,

	"FakeNormalTailTapNote":          0,
	"FakeCriticalTailTapNote":        0,
	"FakeNormalTailFlickNote":        0,
	"FakeCriticalTailFlickNote":      0,
	"FakeNormalTailTraceNote":        0,
	"FakeCriticalTailTraceNote":      0,
	"FakeNormalTailTraceFlickNote":   0,
	"FakeCriticalTailTraceFlickNote": 0,
	"FakeNormalTailReleaseNote":      0,
	"FakeCriticalTailReleaseNote":    0,

	"FakeTransientHiddenTickNote": 0,
	"FakeNormalTickNote":          0,
	"FakeCriticalTickNote":        0,
	"FakeAnchorNote":              0,

	"FakeDamageNote": 0,

	"Skill":       0,
	"FeverChance": 0,
	"FeverStart":  0,
}

func getValueFromData(data []sonolus.LevelDataEntityValue, name string) (float64, error) {
	for _, value := range data {
		if value.Name == name {
			return value.Value, nil
		}
	}
	return 0, fmt.Errorf("value not found: %s", name)
}

func getTimeFromBpmChanges(bpmChanges []BpmChange, beat float64) float64 {
	ret := 0.0
	for i, bpmChange := range bpmChanges {
		if i == len(bpmChanges)-1 {
			ret += (beat - bpmChange.Beat) * (60 / bpmChange.Bpm)
			break
		}
		nextBpmChange := bpmChanges[i+1]
		if beat >= bpmChange.Beat && beat < nextBpmChange.Beat {
			ret += (beat - bpmChange.Beat) * (60 / bpmChange.Bpm)
			break
		} else if beat >= nextBpmChange.Beat {
			ret += (nextBpmChange.Beat - bpmChange.Beat) * (60 / bpmChange.Bpm)
		} else {
			break
		}
	}
	return ret
}

func CheckDamageNote(levelData sonolus.LevelData) bool {
	for _, entity := range levelData.Entities {
		if entity.Archetype == "DamageNote" {
			return true
		}
	}
	return false
}

func CalculateScore(levelInfo sonolus.LevelInfo, levelData sonolus.LevelData, unitScore float64, ignoreDamage bool, listing bool) []YagooFrame {
	var weightedNotesCount float64 = 0
	for _, entity := range levelData.Entities {
		weight := WEIGHT_MAP[entity.Archetype]
		if weight == 0 || (entity.Archetype == "DamageNote" && ignoreDamage) {
			continue
		}
		weightedNotesCount += weight
	}

	frames := make([]YagooFrame, 0, int(weightedNotesCount)+1)
	frames = append(frames, YagooFrame{
		Time:  0,
		Score: 0,
	})
	bpmChanges := ([]BpmChange{})
	comboFax := 1.0

	score := 0.0
	comboCounter := 0
	dataEntities := ([]sonolus.LevelDataEntity{})

	for _, entity := range levelData.Entities {
		weight := WEIGHT_MAP[entity.Archetype]
		if weight > 0.0 && len(entity.Data) > 0 {
			dataEntities = append(dataEntities, entity)
		} else if entity.Archetype == "#BPM_CHANGE" {
			beat, err := getValueFromData(entity.Data, "#BEAT")
			if err != nil {
				continue
			}
			bpm, err := getValueFromData(entity.Data, "#BPM")
			if err != nil {
				continue
			}
			bpmChanges = append(bpmChanges, BpmChange{
				Beat: beat,
				Bpm:  bpm,
			})
		}
	}
	slices.SortStableFunc(dataEntities, func(a, b sonolus.LevelDataEntity) int {
		aBeat, _ := getValueFromData(a.Data, "#BEAT")
		bBeat, _ := getValueFromData(b.Data, "#BEAT")
		if aBeat < bBeat {
			return -1
		}
		if aBeat > bBeat {
			return 1
		}
		return 0
	})
	slices.SortStableFunc(bpmChanges, func(a, b BpmChange) int {
		if a.Beat < b.Beat {
			return -1
		}
		if a.Beat > b.Beat {
			return 1
		}
		return 0
	})

	for _, entity := range dataEntities {
		if listing {
			break
		}
		weight := WEIGHT_MAP[entity.Archetype]

		// get beat/time early so skillFax can be applied to this event
		beat, err := getValueFromData(entity.Data, "#BEAT")
		if err != nil {
			continue
		}
		time := getTimeFromBpmChanges(bpmChanges, beat)

		comboCounter += 1
		if comboCounter%100 == 1 && comboCounter > 1 {
			comboFax += 0.01
		}
		if comboFax > 1.1 {
			comboFax = 1.1
		}

		//	switch scoreMode {
		//	default:
		score += ((float64(unitScore) / weightedNotesCount) *
			//		4 * // Constant
			weight * // Note weight
			//		1 * // Judge weight (Always 1)
			//		levelFax * // Level fax
			comboFax)
		//	case "tournament":
		//		score += 3
		//	}

		frames = append(frames, YagooFrame{
			Time:  time,
			Score: score,
		})
	}

	return frames
}

func WriteYagooFile(frames []YagooFrame, assets string, path string, levelInfo sonolus.LevelInfo, levelData sonolus.LevelData, listing bool) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("ファイルの作成に失敗しました (Failed to create file.) [%s]", err)
	}
	defer file.Close()

	writer := io.Writer(file)

	fmt.Fprintf(writer, "p|%s\n", assets)
	fmt.Fprintf(writer, "b|%s\n", strconv.FormatBool(listing))
	fmt.Fprintf(writer, "v|%s\n", Version)
	fmt.Fprintf(writer, "u|%d\n", time.Now().Unix())

	lastScore := 0.0
	lastRankIndex := 1
	lastRankPlus := 0
	for i, frame := range frames {
		score := frame.Score
		frameScore := math.Round(score - lastScore)
		lastScore = frame.Score

		rankIndex := 1
		rankPlus := 0
		scoreX := 0.0
		scoreXS := 0.0

		const rankS = 1000000.0
		const rankA = 600000.0
		const rankB = 300000.0
		const rankC = 150000.0

		// bar
		const rankPlusPos = 1 // 344.0 / 344.0
		const rankSPos = 1    // 1070.0 / 1070.0
		const rankAPos = 854.0 / 1070.0
		const rankBPos = 636.0 / 1070.0
		const rankCPos = 418.0 / 1070.0

		if score < 0 {
			rankIndex = 1
			rankPlus = 0
			scoreX = 0
			scoreXS = 0
		} else if score >= rankS {
			rankIndex = 5
			scoreX = 1
			const (
				rankSP1     = 1200000.0
				rankSP2     = 1500000.0
				rankSP2Step = 500000.0
			)
			if score < rankSP1 {
				rankPlus = 0
				scoreXS = ((score - rankS) / (rankSP1 - rankS)) * rankPlusPos
			} else if score < rankSP2 {
				rankPlus = 1
				scoreXS = ((score - rankSP1) / (rankSP2 - rankSP1)) * rankPlusPos
			} else {
				rankPlus = 2 + int(math.Floor((score-rankSP2)/rankSP2Step))
				currentThreshold := rankSP2 + float64(rankPlus-2)*rankSP2Step
				scoreXS = ((score - currentThreshold) / rankSP2Step) * rankPlusPos
			}
		} else if score >= rankA {
			rankIndex = 4
			rankPlus = int(math.Round((score - rankA) / (rankS - rankA)))
			scoreX = ((score-rankA)/(rankS-rankA))*(rankSPos-rankAPos) + rankAPos
		} else if score >= rankB {
			rankIndex = 3
			rankPlus = int(math.Round((score - rankB) / (rankA - rankB)))
			scoreX = ((score-rankB)/(rankA-rankB))*(rankAPos-rankBPos) + rankBPos
		} else if score >= rankC {
			rankIndex = 2
			rankPlus = int(math.Round((score - rankC) / (rankB - rankC)))
			scoreX = ((score-rankC)/(rankB-rankC))*(rankBPos-rankCPos) + rankCPos
		} else {
			rankIndex = 1
			rankPlus = int(math.Round(score / rankC))
			scoreX = (score / rankC) * rankCPos
		}

		time := frame.Time
		combo := i

		if i < len(frames)-1 && time == frames[i+1].Time {
			continue
		}
		if time == 0 && i > 0 {
			time = frames[i-1].Time + 0.000001
		}
		if lastRankIndex != rankIndex || lastRankPlus != rankPlus {
			writer.Write(fmt.Appendf(nil, "r|%f:%d:%d\n", time, rankIndex, rankPlus))
			lastRankIndex = rankIndex
			lastRankPlus = rankPlus
		}

		writer.Write(fmt.Appendf(nil, "d|%f:%.0f:%.0f:%f:%f:%d\n", time, score, frameScore, scoreX, scoreXS, combo))
	}
	return nil
}
