package analysis

import (
	"alphabot/models"
	"fmt"
)

type PatternDetectionResult struct {
	Pattern      string
	Confirmation string
	IsDetected   bool
	Direction    int
}

func DetectEngulfing(records []models.AutoVolumeRecord) PatternDetectionResult {
	// records[1] = nến 21
	// records[0] = nến 22 (đã đóng gần nhất)

	if records[1].Candlestick() == 0 &&
		records[0].Candlestick() == 1 &&
		records[0].QuoteAssetVolume > records[1].QuoteAssetVolume*1.2 &&
		records[0].OpenPrice < records[1].ClosePrice &&
		records[0].ClosePrice > records[1].OpenPrice {
		return PatternDetectionResult{
			Pattern:      "⚙️ Mô hình 🐂 Bullish Engulfing",
			Confirmation: "✅ Đây là một tín hiệu đảo chiều tăng giá rất mạnh mẽ, đặc biệt nếu nó xuất hiện sau một xu hướng giảm. Nó cho thấy phe mua đã hoàn toàn áp đảo phe bán",
			IsDetected:   true,
			Direction:    1,
		}
	} else if records[1].Candlestick() == 1 &&
		records[0].Candlestick() == 0 &&
		records[0].QuoteAssetVolume > records[1].QuoteAssetVolume*1.2 &&
		records[0].OpenPrice > records[1].ClosePrice &&
		records[0].ClosePrice < records[1].OpenPrice {
		return PatternDetectionResult{
			Pattern:      "⚙️ Mô hình 🐻 Bearish Engulfing",
			Confirmation: "🍎 Đây là một tín hiệu đảo chiều giảm giá mạnh mẽ, đặc biệt nếu nó xuất hiện sau một xu hướng tăng. Nó cho thấy phe bán đã hoàn toàn áp đảo phe mua",
			IsDetected:   true,
			Direction:    2,
		}
	}
	return PatternDetectionResult{IsDetected: false, Direction: 0}
}

func DetectBreakout(records []models.AutoVolumeRecord, averageCandlestickBody float64) PatternDetectionResult {
	if len(records) < 8 { // Cần ít nhất 8 nến để có nến 15-19
		return PatternDetectionResult{IsDetected: false}
	}

	// QUAN TRỌNG: Phân tích Breakout trên nến ĐÃ ĐÓNG
	// records[0] = nến 22 (đã đóng gần nhất)
	// records[1] = nến 21 (đã đóng trước đó)
	record20 := records[1] // Nến 21
	record21 := records[0] // Nến 22 (đã đóng gần nhất)

	// Tính resistance level (cao nhất của 5 nến trước nến hiện tại)
	resistance := calculateResistance(records)

	if record21.Candlestick() == 1 &&
		record21.IsCandlestickBodyLong(averageCandlestickBody, 1.5) &&
		record21.QuoteAssetVolume > record20.QuoteAssetVolume*1.2 &&
		record20.ClosePrice < resistance && // Nến trước chưa phá vỡ
		record21.ClosePrice > resistance { // Nến hiện tại phá vỡ
		return PatternDetectionResult{
			Pattern:      "⚙️ Mô hình 🐂 Breakout",
			Confirmation: "✅ Tín hiệu breakout: Giá đóng cửa vượt qua resistance",
			IsDetected:   true,
			Direction:    1,
		}
	}
	return PatternDetectionResult{IsDetected: false, Direction: 0}
}

// Tính resistance level (cao nhất của 16 nến trước nến hiện tại)
func calculateResistance(records []models.AutoVolumeRecord) float64 {
	// Kiểm tra điều kiện biên
	if len(records) < 20 { // Cần ít nhất từ records[1] đến records[19]
		return 0
	}
	// Xác định phạm vi nến 3-19 (tương ứng records[19] đến records[3])
	// Vì:
	// records[0] = nến 22 (mới nhất)
	// CORRECTED RANGE: Nến 3-19 tương ứng với records[19] đến records[3]
	startIdx := 19 // nến 3
	endIdx := 3    // nến 19
	if startIdx >= len(records) || endIdx >= len(records) {
		return 0
	}

	resistance := records[startIdx].HighPrice
	for i := startIdx; i >= endIdx; i-- { // Lặp từ nến 3 đến 19
		if records[i].HighPrice > resistance {
			resistance = records[i].HighPrice
		}
	}

	return resistance
}

func DetectHammer(records []models.AutoVolumeRecord) PatternDetectionResult {
	// Kiểm tra điều kiện biên
	if len(records) < 7 {
		return PatternDetectionResult{IsDetected: false, Direction: 0}
	}

	// QUAN TRỌNG: Phân tích Hammer trên nến ĐÃ ĐÓNG (records[0]) - nến 22 (mới nhất đã đóng)
	// records[0] = nến 22 (đã đóng gần nhất) - PHÂN TÍCH Ở ĐÂY
	// records[1] = nến 21 (đã đóng trước đó)
	isDowntrend := checkDowntrendFromIndex(records, 1, 5)
	body := records[0].CandlestickBody()
	totalLength := records[0].CandlestickLength()
	upperShadow := records[0].CandlestickUpperShadow()
	lowerShadow := records[0].CandlestickLowerShadow()

	// KIỂM TRA ĐIỀU KIỆN ĐỂ TRÁNH NaN
	// Nếu totalLength = 0 (HighPrice = LowPrice), không thể phân tích
	if totalLength <= 0 {
		return PatternDetectionResult{IsDetected: false, Direction: 0}
	}
	// Tiêu chuẩn nhận diện Hammer chuyên nghiệp
	validBodySize := body <= totalLength*0.3      // Thân ≤ 30% tổng chiều dài
	validLowerShadow := lowerShadow >= body*2     // Bóng dưới ≥ 2x thân
	minimalUpperShadow := upperShadow <= body*0.5 // Bóng trên ≤ 0.5x thân
	shadowRatio := lowerShadow >= upperShadow*3   // Bóng dưới dài gấp 3x bóng trên
	validPosition := isDowntrend                  // Xuất hiện sau downtrend

	if validBodySize && validLowerShadow && minimalUpperShadow && shadowRatio && validPosition {
		// Phân loại Hammer
		var direction int
		hammerType := "🐂 Bullish"
		confidence := "Tín hiệu mạnh"
		if records[0].ClosePrice < records[0].OpenPrice {
			hammerType = "🐻 Bearish (Hanging Man)"
			confidence = "Cần nến tăng xác nhận"
			direction = 2
		} else {
			direction = 1
		}

		return PatternDetectionResult{
			Pattern: fmt.Sprintf("⚙️ Mô hình Hammer (%s)", hammerType),
			Confirmation: fmt.Sprintf("✅ %s - Thân: %.2f%%, Bóng dưới: %.2f%%, Bóng trên: %.2f%%",
				confidence,
				(body/totalLength)*100,
				(lowerShadow/totalLength)*100,
				(upperShadow/totalLength)*100),
			IsDetected: true,
			Direction:  direction,
		}
	}
	return PatternDetectionResult{IsDetected: false, Direction: 0}
}

// Hàm kiểm tra downtrend từ index cụ thể
func checkDowntrendFromIndex(records []models.AutoVolumeRecord, startIdx, endIdx int) bool {
	// Kiểm tra điều kiện biên
	if startIdx >= len(records) || endIdx >= len(records) || startIdx > endIdx {
		return false
	}

	// Tính số lượng nến giảm trong khoảng từ startIdx đến endIdx
	downCount := 0
	totalCandles := endIdx - startIdx + 1

	for i := startIdx; i <= endIdx; i++ {
		if records[i].Candlestick() == 0 { // Nến giảm
			downCount++
		}
	}

	// Xác định xu hướng giảm (ít nhất 60% nến là giảm)
	return float64(downCount)/float64(totalCandles) >= 0.6
}

func DetectDojiSpecial(records []models.AutoVolumeRecord) PatternDetectionResult {
	const (
		bodyThreshold   = 0.1  // Thân nến ≤ 10% tổng độ dài
		shadowThreshold = 0.05 // Bóng ngắn ≤ 5% tổng độ dài
		minShadowRatio  = 2.0  // Bóng dài phải gấp ít nhất 2 lần thân
	)

	// Kiểm tra điều kiện biên - cần ít nhất 7 nến để phân tích
	if len(records) < 7 {
		return PatternDetectionResult{IsDetected: false, Direction: 0}
	}

	// QUAN TRỌNG: Phân tích Doji trên nến ĐÃ ĐÓNG (records[0]) - nến 22 (mới nhất đã đóng)
	// records[0] = nến 22 (đã đóng gần nhất) - PHÂN TÍCH Ở ĐÂY
	// records[1] = nến 21 (đã đóng trước đó)
	candle := records[0] // Nến 22 (đã đóng gần nhất)
	body := candle.CandlestickBody()
	totalLength := candle.CandlestickLength()
	upperShadow := candle.CandlestickUpperShadow()
	lowerShadow := candle.CandlestickLowerShadow()

	// KIỂM TRA ĐIỀU KIỆN ĐỂ TRÁNH NaN
	// Nếu totalLength = 0 (HighPrice = LowPrice), không thể phân tích
	if totalLength <= 0 {
		return PatternDetectionResult{IsDetected: false, Direction: 0}
	}

	// Nếu body = 0, không thể tính tỷ lệ
	if body <= 0 {
		return PatternDetectionResult{IsDetected: false, Direction: 0}
	}

	// Bỏ qua nếu không phải Doji (thân quá lớn)
	if body > totalLength*bodyThreshold {
		return PatternDetectionResult{IsDetected: false, Direction: 0}
	}

	// Kiểm tra xu hướng trước đó (5 nến trước records[0])
	// Sử dụng records[1] đến records[5] để kiểm tra xu hướng
	isDowntrend := checkDowntrendFromIndex(records, 1, 5)
	isUptrend := checkUptrendFromIndex(records, 1, 5)

	// Kiểm tra Dragonfly Doji (bóng dưới dài, bóng trên rất ngắn)
	if upperShadow <= totalLength*shadowThreshold &&
		lowerShadow >= totalLength*0.3 && // Bóng dưới ≥ 30% tổng độ dài
		lowerShadow >= body*minShadowRatio {

		// Dragonfly Doji chỉ có ý nghĩa sau downtrend
		if isDowntrend {
			return PatternDetectionResult{
				Pattern:      "⚙️ Mô hình 🐂 Dragonfly Doji",
				Confirmation: "✅ Tín hiệu tăng mạnh sau downtrend - Phe mua đang tích lũy",
				IsDetected:   true,
				Direction:    1,
			}
		} else {
			return PatternDetectionResult{
				Pattern:      "⚙️ Mô hình 🐂 Dragonfly Doji (Yếu)",
				Confirmation: "⚠️ Dragonfly Doji xuất hiện không sau downtrend - Tín hiệu yếu hơn",
				IsDetected:   true,
				Direction:    1,
			}
		}
	}

	// Kiểm tra Gravestone Doji (bóng trên dài, bóng dưới rất ngắn)
	if lowerShadow <= totalLength*shadowThreshold &&
		upperShadow >= totalLength*0.3 && // Bóng trên ≥ 30% tổng độ dài
		upperShadow >= body*minShadowRatio {

		// Gravestone Doji chỉ có ý nghĩa sau uptrend
		if isUptrend {
			return PatternDetectionResult{
				Pattern:      "⚙️ Mô hình 🐻 Gravestone Doji",
				Confirmation: "🍎 Tín hiệu giảm mạnh sau uptrend - Phe bán đang áp đảo",
				IsDetected:   true,
				Direction:    2,
			}
		} else {
			return PatternDetectionResult{
				Pattern:      "⚙️ Mô hình 🐻 Gravestone Doji (Yếu)",
				Confirmation: "⚠️ Gravestone Doji xuất hiện không sau uptrend - Tín hiệu yếu hơn",
				IsDetected:   true,
				Direction:    2,
			}
		}
	}

	// Kiểm tra Four Price Doji (thân rất nhỏ, bóng trên và dưới đều ngắn)
	if body <= totalLength*0.05 && // Thân ≤ 5% tổng độ dài
		upperShadow <= totalLength*0.1 && // Bóng trên ≤ 10%
		lowerShadow <= totalLength*0.1 { // Bóng dưới ≤ 10%

		return PatternDetectionResult{
			Pattern:      "⚙️ Mô hình 🔄 Four Price Doji",
			Confirmation: "🔄 Tín hiệu indecision - Thị trường đang cân bằng, chờ breakout",
			IsDetected:   true,
			Direction:    0, // Không xác định hướng
		}
	}

	return PatternDetectionResult{IsDetected: false, Direction: 0}
}

// Hàm kiểm tra uptrend từ index cụ thể
func checkUptrendFromIndex(records []models.AutoVolumeRecord, startIdx, endIdx int) bool {
	// Kiểm tra điều kiện biên
	if startIdx >= len(records) || endIdx >= len(records) || startIdx > endIdx {
		return false
	}

	// Tính số lượng nến tăng trong khoảng từ startIdx đến endIdx
	upCount := 0
	totalCandles := endIdx - startIdx + 1

	for i := startIdx; i <= endIdx; i++ {
		if records[i].Candlestick() == 1 { // Nến tăng
			upCount++
		}
	}

	// Xác định xu hướng tăng (ít nhất 60% nến là tăng)
	return float64(upCount)/float64(totalCandles) >= 0.6
}
