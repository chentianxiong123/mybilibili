package com.mybilibili.gateway.filter;

import com.mybilibili.gateway.security.GatewayUserPrincipal;
import org.springframework.cloud.gateway.filter.GatewayFilterChain;
import org.springframework.cloud.gateway.filter.GlobalFilter;
import org.springframework.core.Ordered;
import org.springframework.http.HttpHeaders;
import org.springframework.http.server.reactive.ServerHttpRequest;
import org.springframework.security.core.Authentication;
import org.springframework.stereotype.Component;
import org.springframework.web.server.ServerWebExchange;
import reactor.core.publisher.Mono;

@Component
public class SecurityUserHeaderFilter implements GlobalFilter, Ordered {

    @Override
    public Mono<Void> filter(ServerWebExchange exchange, GatewayFilterChain chain) {
        ServerHttpRequest.Builder requestBuilder = exchange.getRequest().mutate()
                .headers(headers -> {
                    headers.remove("X-User-Id");
                    headers.remove("X-Admin-Id");
                    headers.remove("X-Username");
                    headers.remove("X-User-Role");
                });

        return exchange.getPrincipal()
                .cast(Authentication.class)
                .filter(Authentication::isAuthenticated)
                .filter(authentication -> authentication.getPrincipal() instanceof GatewayUserPrincipal)
                .map(authentication -> {
                    GatewayUserPrincipal principal = (GatewayUserPrincipal) authentication.getPrincipal();
                    return requestBuilder.headers(headers -> {
                        headers.set("X-User-Id", String.valueOf(principal.userId()));
                        headers.set("X-Username", principal.username() == null ? "" : principal.username());
                        headers.set("X-User-Role", principal.role() == null ? "USER" : principal.role());
                        if (principal.isAdmin()) {
                            headers.set("X-Admin-Id", String.valueOf(principal.userId()));
                        }
                        String authHeader = exchange.getRequest().getHeaders().getFirst(HttpHeaders.AUTHORIZATION);
                        if (authHeader != null) {
                            headers.set(HttpHeaders.AUTHORIZATION, authHeader);
                        }
                    }).build();
                })
                .defaultIfEmpty(requestBuilder.build())
                .flatMap(request -> chain.filter(exchange.mutate().request(request).build()));
    }

    @Override
    public int getOrder() {
        return -50;
    }
}
