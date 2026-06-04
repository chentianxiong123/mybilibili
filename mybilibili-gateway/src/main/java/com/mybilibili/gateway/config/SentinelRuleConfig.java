package com.mybilibili.gateway.config;

import com.alibaba.csp.sentinel.adapter.gateway.common.SentinelGatewayConstants;
import com.alibaba.csp.sentinel.adapter.gateway.common.api.ApiDefinition;
import com.alibaba.csp.sentinel.adapter.gateway.common.api.ApiPathPredicateItem;
import com.alibaba.csp.sentinel.adapter.gateway.common.api.ApiPredicateItem;
import com.alibaba.csp.sentinel.adapter.gateway.common.api.GatewayApiDefinitionManager;
import com.alibaba.csp.sentinel.adapter.gateway.common.rule.GatewayFlowRule;
import com.alibaba.csp.sentinel.adapter.gateway.common.rule.GatewayRuleManager;
import jakarta.annotation.PostConstruct;
import org.springframework.context.annotation.Configuration;

import java.util.HashSet;
import java.util.Set;

@Configuration
public class SentinelRuleConfig {

    @PostConstruct
    public void initRules() {
        initApiDefinitions();
        initFlowRules();
    }

    private void initApiDefinitions() {
        Set<ApiDefinition> definitions = new HashSet<>();

        definitions.add(new ApiDefinition("video-api")
                .setPredicateItems(Set.of(
                        new ApiPathPredicateItem().setPattern("/api/video/**")
                                .setMatchStrategy(SentinelGatewayConstants.URL_MATCH_STRATEGY_PREFIX),
                        new ApiPathPredicateItem().setPattern("/api/manuscript/**")
                                .setMatchStrategy(SentinelGatewayConstants.URL_MATCH_STRATEGY_PREFIX)
                )));

        definitions.add(new ApiDefinition("creator-api")
                .setPredicateItems(Set.of(
                        new ApiPathPredicateItem().setPattern("/api/creator/**")
                                .setMatchStrategy(SentinelGatewayConstants.URL_MATCH_STRATEGY_PREFIX)
                )));

        definitions.add(new ApiDefinition("user-api")
                .setPredicateItems(Set.of(
                        new ApiPathPredicateItem().setPattern("/api/user/**")
                                .setMatchStrategy(SentinelGatewayConstants.URL_MATCH_STRATEGY_PREFIX)
                )));

        definitions.add(new ApiDefinition("search-api")
                .setPredicateItems(Set.of(
                        new ApiPathPredicateItem().setPattern("/api/search/**")
                                .setMatchStrategy(SentinelGatewayConstants.URL_MATCH_STRATEGY_PREFIX)
                )));

        definitions.add(new ApiDefinition("interaction-api")
                .setPredicateItems(Set.of(
                        new ApiPathPredicateItem().setPattern("/api/comment/**")
                                .setMatchStrategy(SentinelGatewayConstants.URL_MATCH_STRATEGY_PREFIX),
                        new ApiPathPredicateItem().setPattern("/api/danmaku/**")
                                .setMatchStrategy(SentinelGatewayConstants.URL_MATCH_STRATEGY_PREFIX)
                )));

        definitions.add(new ApiDefinition("auth-api")
                .setPredicateItems(Set.of(
                        new ApiPathPredicateItem().setPattern("/api/user/login")
                                .setMatchStrategy(SentinelGatewayConstants.URL_MATCH_STRATEGY_EXACT),
                        new ApiPathPredicateItem().setPattern("/api/user/register")
                                .setMatchStrategy(SentinelGatewayConstants.URL_MATCH_STRATEGY_EXACT)
                )));

        GatewayApiDefinitionManager.loadApiDefinitions(definitions);
    }

    private void initFlowRules() {
        Set<GatewayFlowRule> rules = new HashSet<>();

        // 全局限流: 每秒 200 请求
        rules.add(new GatewayFlowRule("video-api")
                .setResourceMode(SentinelGatewayConstants.RESOURCE_MODE_CUSTOM_API_NAME)
                .setCount(200)
                .setIntervalSec(1));

        rules.add(new GatewayFlowRule("creator-api")
                .setResourceMode(SentinelGatewayConstants.RESOURCE_MODE_CUSTOM_API_NAME)
                .setCount(80)
                .setIntervalSec(1));

        rules.add(new GatewayFlowRule("user-api")
                .setResourceMode(SentinelGatewayConstants.RESOURCE_MODE_CUSTOM_API_NAME)
                .setCount(100)
                .setIntervalSec(1));

        rules.add(new GatewayFlowRule("search-api")
                .setResourceMode(SentinelGatewayConstants.RESOURCE_MODE_CUSTOM_API_NAME)
                .setCount(100)
                .setIntervalSec(1));

        rules.add(new GatewayFlowRule("interaction-api")
                .setResourceMode(SentinelGatewayConstants.RESOURCE_MODE_CUSTOM_API_NAME)
                .setCount(150)
                .setIntervalSec(1));

        // 登录/注册接口严格限流: 每秒 10 次（防暴力破解）
        rules.add(new GatewayFlowRule("auth-api")
                .setResourceMode(SentinelGatewayConstants.RESOURCE_MODE_CUSTOM_API_NAME)
                .setCount(10)
                .setIntervalSec(1));

        GatewayRuleManager.loadRules(rules);
    }
}
