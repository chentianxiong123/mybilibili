package com.mybilibili.gateway.security;

import com.mybilibili.gateway.filter.GatewayRequestPolicy;
import org.springframework.security.authorization.AuthorizationDecision;
import org.springframework.security.authorization.ReactiveAuthorizationManager;
import org.springframework.security.core.Authentication;
import org.springframework.security.web.server.authorization.AuthorizationContext;
import org.springframework.stereotype.Component;
import reactor.core.publisher.Mono;

@Component
public class GatewayAuthorizationManager implements ReactiveAuthorizationManager<AuthorizationContext> {

    private final GatewayRequestPolicy requestPolicy;

    public GatewayAuthorizationManager(GatewayRequestPolicy requestPolicy) {
        this.requestPolicy = requestPolicy;
    }

    @Override
    public Mono<AuthorizationDecision> check(Mono<Authentication> authentication, AuthorizationContext context) {
        String path = context.getExchange().getRequest().getURI().getPath();
        var method = context.getExchange().getRequest().getMethod();
        if (requestPolicy.isPublicPath(method, path)) {
            return Mono.just(new AuthorizationDecision(true));
        }

        return authentication
                .filter(Authentication::isAuthenticated)
                .map(auth -> new AuthorizationDecision(isAllowed(auth, method, path)))
                .defaultIfEmpty(new AuthorizationDecision(false));
    }

    private boolean isAllowed(Authentication authentication, org.springframework.http.HttpMethod method, String path) {
        if (requestPolicy.isSuperAdminPath(method, path)) {
            return hasAuthority(authentication, "ROLE_SUPER_ADMIN");
        }
        if (requestPolicy.isAdminPath(method, path)) {
            if (!hasAuthority(authentication, "ROLE_ADMIN")) {
                return false;
            }
            if (hasAuthority(authentication, "ROLE_SUPER_ADMIN")) {
                return true;
            }
            String requiredPermission = requestPolicy.requiredPermission(method, path);
            return requiredPermission != null
                    && hasAuthority(authentication, "PERMISSION_" + requiredPermission);
        }
        return true;
    }

    private boolean hasAuthority(Authentication authentication, String authority) {
        return authentication.getAuthorities().stream()
                .anyMatch(grantedAuthority -> authority.equals(grantedAuthority.getAuthority()));
    }
}
