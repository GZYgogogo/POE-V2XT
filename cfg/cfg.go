package cfg

type Config struct {
	Constant *Constant
	HardWare *HardWare
}

type Constant struct {
	Alpha        float64 `json:"alpha"`
	Beta         float64 `json:"beta"`
	Gamma        float64 `json:"gamma"`
	Delta        float64 `json:"delta"`
	Fai          float64 `json:"fai"`
	Fee          float64 `json:"fee"`           // 手续费
	Kappa        float64 `json:"kappa"`         // 惩罚因子
	Vs           float64 `json:"vs"`            // 提交奖励因子
	Vv           float64 `json:"vv"`            // 验证奖励因子
	Theta        float64 `json:"theta"`         // 成本预订范围最大值
	Tau          float64 `json:"tau"`           // 交易提交到最终保存在链上的最大可容忍延迟，单位为秒
	ExpectedTime float64 `json:"expected_time"` // 期望交易验证完成时间，单位为秒
}

type HardWare struct {
	InitCredit float64 `json:"civ"`
	Wcpu       float64 `json:"wcpu"`   // CPU因子
	Wmem       float64 `json:"wmem"`   // 内存因子
	Whard      float64 `json:"whard"`  // 硬盘因子
	Bmcpu      float64 `json:"bmcpu"`  // CPU基准硬件值
	Bmmem      float64 `json:"bmmem"`  // 内存基准硬件值
	Bmhard     float64 `json:"bmhard"` // 硬盘基准硬件值
}

var Cfg *Config

func IninDefault() {
	Cfg = &Config{}
}

func Init() *Config {
	IninDefault()
	return Cfg
}
