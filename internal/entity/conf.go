package entity

// AppConf 应用配置
type AppConf struct {
	PrometheusCfg *PrometheusConf `yaml:"prometheusCfg"` // Prometheus 配置
}

// PrometheusConf Prometheus 配置
type PrometheusConf struct {
	Enable  bool   `yaml:"enable"`
	Port    int    `yaml:"port"`
	Address string `yaml:"address"`
}
