-- AI channel provider-specific extension config.
ALTER TABLE `ai_api_configs`
  ADD COLUMN `extra_config` json NULL COMMENT 'Provider-specific JSON config' AFTER `enabled`;
