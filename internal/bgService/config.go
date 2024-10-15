package bgService

import (
	"fmt"
	"win-sense-connect/internal/common"
)

func (p *program) loadConfig(logger common.Logger) error {
	logger.Debug("Starting to load config...")
	conf, err := p.db.GetConfig()
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get config: %v", err))
		return err
	}
	p.config = *conf
	logger.Debug("Config loaded successfully")
	return nil
}
