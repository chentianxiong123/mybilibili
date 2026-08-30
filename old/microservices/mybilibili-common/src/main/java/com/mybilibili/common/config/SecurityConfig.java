package com.mybilibili.common.config;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.security.authorization.AuthorizationDecision;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.config.annotation.web.configurers.AbstractHttpConfigurer;
import org.springframework.security.config.http.SessionCreationPolicy;
import org.springframework.security.core.Authentication;
import org.springframework.security.web.SecurityFilterChain;
import org.springframework.security.web.access.intercept.RequestAuthorizationContext;
import org.springframework.security.web.authentication.UsernamePasswordAuthenticationFilter;

import java.util.function.Supplier;

@Configuration
public class SecurityConfig {

    @Autowired
    private JwtFilter jwtFilter;

    private final JwtRequestPolicy requestPolicy = new JwtRequestPolicy();

    @Bean
    public SecurityFilterChain securityFilterChain(HttpSecurity http) throws Exception {
        return http
                .csrf(AbstractHttpConfigurer::disable)
                .httpBasic(AbstractHttpConfigurer::disable)
                .formLogin(AbstractHttpConfigurer::disable)
                .logout(AbstractHttpConfigurer::disable)
                .sessionManagement(session -> session.sessionCreationPolicy(SessionCreationPolicy.STATELESS))
                .authorizeHttpRequests(authorize -> authorize
                        .requestMatchers(request -> requestPolicy.isPublicPath(request.getMethod(), request.getRequestURI())).permitAll()
                        .anyRequest().access(this::authorize))
                .addFilterBefore(jwtFilter, UsernamePasswordAuthenticationFilter.class)
                .build();
    }

    private AuthorizationDecision authorize(Supplier<Authentication> authentication,
                                            RequestAuthorizationContext context) {
        Authentication current = authentication.get();
        if (current == null || !current.isAuthenticated()) {
            return new AuthorizationDecision(false);
        }

        var request = context.getRequest();
        String method = request.getMethod();
        String path = request.getRequestURI();
        if (requestPolicy.isSuperAdminPath(method, path)) {
            return new AuthorizationDecision(hasAuthority(current, "ROLE_SUPER_ADMIN"));
        }
        if (requestPolicy.isAdminPath(method, path)) {
            String requiredPermission = requestPolicy.requiredPermission(method, path);
            boolean allowed = hasAuthority(current, "ROLE_ADMIN")
                    && (hasAuthority(current, "ROLE_SUPER_ADMIN")
                    || requiredPermission != null && hasAuthority(current, "PERMISSION_" + requiredPermission));
            return new AuthorizationDecision(allowed);
        }
        return new AuthorizationDecision(true);
    }

    private boolean hasAuthority(Authentication authentication, String authority) {
        return authentication.getAuthorities().stream()
                .anyMatch(grantedAuthority -> authority.equals(grantedAuthority.getAuthority()));
    }
}
