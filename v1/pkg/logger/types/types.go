package types

type LoggerOption struct {
	EnableSaveFile bool   `yaml:"enable_save_file" mapstructure:"enable_save_file"`
	SavePath       string `yaml:"save_path" mapstructure:"save_path"`

	MaxBackups int `yaml:"max_backups" mapstructure:"max_backups"`
	MaxSize    int `yaml:"max_size" mapstructure:"max_size"`
	MaxAge     int `yaml:"max_age" mapstructure:"max_age"`
}
