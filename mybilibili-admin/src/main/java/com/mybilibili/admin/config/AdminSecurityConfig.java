package com.mybilibili.admin.config;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.annotation.Configuration;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.config.annotation.web.configuration.EnableWebSecurity;
import org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter;
import org.springframework.security.web.authentication.UsernamePasswordAuthenticationFilter;

@Configuration
@EnableWebSecurity
public class AdminSecurityConfig extends WebSecurityConfigurerAdapter {

    @Autowired
    private AdminJwtFilter adminJwtFilter;

    @Override
    protected void configure(HttpSecurity http) throws Exception {
        http
            .csrf().disable()
            .authorizeRequests()
            .antMatchers("/user/login", "/user/register", "/test", "/swagger-ui.html", "/swagger-ui/**", "/v3/api-docs/**", "/swagger-resources/**", "/webjars/**").permitAll()
            .antMatchers("/uploads/**", "/admin/uploads/**").permitAll()
            .antMatchers("/manuscript/test-ai-api", "/manuscript/test-ai-summary/**").permitAll()
            // 暂时取消上架下架接口的权限校验
            .antMatchers("/manuscript/publish/**", "/manuscript/unpublish/**").permitAll()
            .anyRequest().authenticated();

        http.addFilterBefore(adminJwtFilter, UsernamePasswordAuthenticationFilter.class);
    }
}