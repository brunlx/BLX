package com.brunlx.blx;

import android.app.Activity;
import android.content.Context;
import android.graphics.Color;
import android.os.Bundle;
import android.util.Log;
import android.view.KeyEvent;
import android.view.View;
import android.view.Window;
import android.view.WindowManager;
import android.webkit.WebResourceRequest;
import android.webkit.WebSettings;
import android.webkit.WebView;
import android.webkit.WebViewClient;
import android.widget.FrameLayout;
import android.widget.ProgressBar;
import android.widget.Toast;

import java.io.File;
import java.io.IOException;
import java.net.Socket;

public class MainActivity extends Activity {
    private static final String TAG = "BLX";
    private static final int PORT = 8080;
    private static final String BASE_URL = "http://127.0.0.1:" + PORT + "/";

    private WebView web;
    private ProgressBar spinner;
    private Process server;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setupUI();
        startServer();
    }

    private void setupUI() {
        Window w = getWindow();
        w.setStatusBarColor(Color.rgb(4, 11, 24));
        w.setNavigationBarColor(Color.rgb(4, 11, 24));

        FrameLayout root = new FrameLayout(this);
        root.setBackgroundColor(Color.rgb(4, 11, 24));

        web = new WebView(this);
        FrameLayout.LayoutParams wl = new FrameLayout.LayoutParams(
                FrameLayout.LayoutParams.MATCH_PARENT,
                FrameLayout.LayoutParams.MATCH_PARENT);
        root.addView(web, wl);

        spinner = new ProgressBar(this, null, android.R.attr.progressBarStyleLarge);
        spinner.getIndeterminateDrawable().setTint(Color.rgb(34, 211, 238));
        FrameLayout.LayoutParams sl = new FrameLayout.LayoutParams(
                FrameLayout.LayoutParams.WRAP_CONTENT,
                FrameLayout.LayoutParams.WRAP_CONTENT);
        sl.gravity = android.view.Gravity.CENTER;
        root.addView(spinner, sl);
        spinner.setVisibility(View.VISIBLE);

        setContentView(root);

        WebSettings s = web.getSettings();
        s.setJavaScriptEnabled(true);
        s.setDomStorageEnabled(true);
        s.setCacheMode(WebSettings.LOAD_NO_CACHE);
        web.setWebViewClient(new WebViewClient() {
            @Override
            public boolean shouldOverrideUrlLoading(WebView view, WebResourceRequest req) {
                String host = req.getUrl().getHost();
                return !(host != null && host.equals("127.0.0.1"));
            }

            @Override
            public void onPageFinished(WebView view, String url) {
                spinner.setVisibility(View.GONE);
            }
        });
    }

    private void startServer() {
        final File bin = resolveBinary();
        if (bin == null) {
            fail("Servidor BLX ausente neste aparelho (arquitetura nao suportada).");
            return;
        }
        if (!bin.canExecute()) {
            Log.w(TAG, "binario sem permissao de execucao: " + bin);
        }

        File log = new File(getFilesDir(), "blx-server.log");
        try {
            ProcessBuilder pb = new ProcessBuilder(bin.getAbsolutePath(), "-no-browser");
            pb.environment().put("HOST", "127.0.0.1");
            pb.environment().put("PORT", String.valueOf(PORT));
            pb.redirectErrorStream(true);
            pb.redirectOutput(log);
            server = pb.start();
        } catch (IOException e) {
            Log.e(TAG, "falha ao iniciar servidor", e);
            fail("Falha ao iniciar o servidor: " + e.getMessage());
            return;
        }

        Thread t = new Thread(new Runnable() {
            @Override
            public void run() {
                long deadline = System.currentTimeMillis() + 15000;
                while (System.currentTimeMillis() < deadline) {
                    if (server != null && !server.isAlive()) {
                        postFail("Servidor encerrou inesperadamente (veja o log).");
                        return;
                    }
                    if (ping()) {
                        runOnUiThread(new Runnable() {
                            @Override
                            public void run() {
                                spinner.setVisibility(View.GONE);
                                web.loadUrl(BASE_URL);
                            }
                        });
                        return;
                    }
                    try {
                        Thread.sleep(200);
                    } catch (InterruptedException ignored) {
                        return;
                    }
                }
                postFail("O servidor BLX demorou a responder.");
            }
        });
        t.setDaemon(true);
        t.start();
    }

    /**
     * O instalador extrai libblxserver.so (a partir de jniLibs/<abi>/ no APK)
     * para nativeLibraryDir, um diretorio marcado como exec_type pelo SELinux
     * (executar binarios a partir de getFilesDir() e bloqueado desde o
     * Android 10). Resolve o caminho real da lib nativa.
     */
    private File resolveBinary() {
        String libDir = getApplicationInfo().nativeLibraryDir;
        File[] candidates = new File[]{
                new File(libDir, "libblxserver.so"),
                new File(new File(libDir, "lib"), "libblxserver.so"),
        };
        for (File c : candidates) {
            if (c.exists()) {
                return c;
            }
        }
        return null;
    }

    private boolean ping() {
        try {
            Socket s = new Socket("127.0.0.1", PORT);
            s.close();
            return true;
        } catch (IOException e) {
            return false;
        }
    }

    private void postFail(final String msg) {
        runOnUiThread(new Runnable() {
            @Override
            public void run() {
                spinner.setVisibility(View.GONE);
                Toast.makeText(MainActivity.this, msg, Toast.LENGTH_LONG).show();
            }
        });
    }

    private void fail(String msg) {
        spinner.setVisibility(View.GONE);
        Toast.makeText(this, msg, Toast.LENGTH_LONG).show();
    }

    @Override
    public boolean onKeyDown(int keyCode, KeyEvent event) {
        if (keyCode == KeyEvent.KEYCODE_BACK && web.canGoBack()) {
            web.goBack();
            return true;
        }
        return super.onKeyDown(keyCode, event);
    }

    @Override
    protected void onDestroy() {
        if (server != null) {
            server.destroy();
            try {
                server.waitFor(1, java.util.concurrent.TimeUnit.SECONDS);
            } catch (InterruptedException ignored) {
            }
            if (server.isAlive()) {
                server.destroyForcibly();
            }
        }
        super.onDestroy();
    }
}
