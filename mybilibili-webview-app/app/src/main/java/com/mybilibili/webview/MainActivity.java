package com.mybilibili.webview;

import android.annotation.SuppressLint;
import android.graphics.Bitmap;
import android.os.Bundle;
import android.view.View;
import android.webkit.WebChromeClient;
import android.webkit.WebSettings;
import android.webkit.WebView;
import android.webkit.WebViewClient;
import android.widget.ProgressBar;

import androidx.appcompat.app.AppCompatActivity;

public class MainActivity extends AppCompatActivity {

    private static final String URL = "http://192.168.31.204:5174/wap/";

    private WebView webView;
    private ProgressBar progressBar;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);

        webView = findViewById(R.id.webView);
        progressBar = findViewById(R.id.progressBar);

        NativeStorage nativeStorage = new NativeStorage(getApplicationContext());
        webView.addJavascriptInterface(nativeStorage, "NativeStorage");

        setupWebView();
        webView.loadUrl(URL);
    }

    @SuppressLint("SetJavaScriptEnabled")
    private void setupWebView() {
        WebSettings settings = webView.getSettings();
        settings.setJavaScriptEnabled(true);
        settings.setDomStorageEnabled(true);
        settings.setDatabaseEnabled(true);
        settings.setCacheMode(WebSettings.LOAD_DEFAULT);
        // wap 页面已自带 viewport meta（width=device-width），按移动端宽度渲染即可。
        // 不要开启 useWideViewPort / loadWithOverviewMode：它们会让 WebView 以宽视口/整页
        // 方式加载后再缩放适配屏幕，加载时产生从左上角往右下角扩散的缩放动画（每次打开都闪）。
        settings.setUseWideViewPort(false);
        settings.setLoadWithOverviewMode(false);
        // 页面已禁用用户缩放（maximum-scale=1），无需内建缩放控件
        settings.setBuiltInZoomControls(false);
        settings.setDisplayZoomControls(false);
        settings.setSupportZoom(false);
        settings.setAllowFileAccess(true);
        settings.setAllowContentAccess(true);
        settings.setMixedContentMode(WebSettings.MIXED_CONTENT_ALWAYS_ALLOW);
        settings.setMediaPlaybackRequiresUserGesture(false);

        webView.setWebViewClient(new WebViewClient() {
            @Override
            public void onPageStarted(WebView view, String url, Bitmap favicon) {
                super.onPageStarted(view, url, favicon);
                progressBar.setVisibility(View.VISIBLE);
            }

            @Override
            public void onPageFinished(WebView view, String url) {
                super.onPageFinished(view, url);
                progressBar.setVisibility(View.GONE);
                verifyBridge();
            }

            private void verifyBridge() {
                // JS 返回测试结果字符串；直接在页面执行处输出到 logcat
                webView.evaluateJavascript(
                        "(function(){try{if(!window.NativeStorage)return 'nobridge';" +
                                "window.NativeStorage.set('bridge:test', Date.now().toString());" +
                                "var v=window.NativeStorage.get('bridge:test');" +
                                "window.NativeStorage.remove('bridge:test');" +
                                "return v?('bridge-ok:'+v):'bridge-empty';" +
                                "}catch(e){return 'bridge-err:'+(e&&e.message);}})();",
                        value -> android.util.Log.i("MyBilibiliBridge",
                                "NATIVE_STORAGE_TEST=" + (value != null ? value : "null")));
            }

            @Override
            public boolean shouldOverrideUrlLoading(WebView view, String url) {
                view.loadUrl(url);
                return true;
            }
        });

        webView.setWebChromeClient(new WebChromeClient() {
            @Override
            public void onProgressChanged(WebView view, int newProgress) {
                super.onProgressChanged(view, newProgress);
                progressBar.setProgress(newProgress);
            }
        });
    }

    @Override
    public void onBackPressed() {
        if (webView.canGoBack()) {
            webView.goBack();
        } else {
            super.onBackPressed();
        }
    }
}