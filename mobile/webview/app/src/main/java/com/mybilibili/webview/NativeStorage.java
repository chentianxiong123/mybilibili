package com.mybilibili.webview;

import android.content.ContentValues;
import android.content.Context;
import android.database.Cursor;
import android.database.sqlite.SQLiteDatabase;
import android.database.sqlite.SQLiteOpenHelper;
import android.webkit.JavascriptInterface;

import java.util.ArrayList;
import java.util.List;

import org.json.JSONArray;
import org.json.JSONObject;

/** KV 型本地存储，数据放 app 私有 databases/ 目录，不受 WebView 缓存清理影响。 */
public class NativeStorage extends SQLiteOpenHelper {

    private static final String DB_NAME = "mybilibili.db";
    private static final String TABLE = "kv";
    private static final String COL_KEY = "key";
    private static final String COL_VALUE = "value";
    private static final String COL_UPDATED = "updated_at";

    public NativeStorage(Context context) {
        super(context, DB_NAME, null, 1);
    }

    @Override
    public void onCreate(SQLiteDatabase db) {
        db.execSQL("CREATE TABLE IF NOT EXISTS " + TABLE + " ("
                + COL_KEY + " TEXT PRIMARY KEY, "
                + COL_VALUE + " TEXT NOT NULL, "
                + COL_UPDATED + " INTEGER NOT NULL)");
    }

    @Override
    public void onUpgrade(SQLiteDatabase db, int oldV, int newV) {
        db.execSQL("DROP TABLE IF EXISTS " + TABLE);
        onCreate(db);
    }

    private SQLiteDatabase w() {
        return getWritableDatabase();
    }

    @JavascriptInterface
    public String get(String key) {
        try (Cursor c = w().query(TABLE, new String[]{COL_VALUE},
                COL_KEY + "=?", new String[]{key}, null, null, null)) {
            if (c != null && c.moveToFirst()) {
                return c.getString(0);
            }
        }
        return null;
    }

    @JavascriptInterface
    public void set(String key, String value) {
        ContentValues v = new ContentValues();
        v.put(COL_KEY, key);
        v.put(COL_VALUE, value);
        v.put(COL_UPDATED, System.currentTimeMillis());
        w().insertWithOnConflict(TABLE, null, v, SQLiteDatabase.CONFLICT_REPLACE);
    }

    @JavascriptInterface
    public void remove(String key) {
        w().delete(TABLE, COL_KEY + "=?", new String[]{key});
    }

    /** 返回 [{key, value}] 的 JSON 数组串 */
    @JavascriptInterface
    public String multiGet(String prefix) {
        try {
            List<JSONObject> out = new ArrayList<>();
            try (Cursor c = w().query(TABLE, new String[]{COL_KEY, COL_VALUE},
                    COL_KEY + " LIKE ?", new String[]{prefix + "%"}, null, null, null)) {
                while (c != null && c.moveToNext()) {
                    JSONObject o = new JSONObject();
                    o.put("key", c.getString(0));
                    o.put("value", c.getString(1));
                    out.add(o);
                }
            }
            JSONArray arr = new JSONArray();
            for (JSONObject o : out) arr.put(o);
            return arr.toString();
        } catch (Exception e) {
            return "[]";
        }
    }

    /** 批量写入，接收 [{key, value}] JSON 数组串 */
    @JavascriptInterface
    public void bulkSet(String json) {
        SQLiteDatabase db = w();
        db.beginTransaction();
        try {
            JSONArray arr = new JSONArray(json);
            for (int i = 0; i < arr.length(); i++) {
                JSONObject o = arr.getJSONObject(i);
                ContentValues v = new ContentValues();
                v.put(COL_KEY, o.optString("key"));
                v.put(COL_VALUE, o.optString("value"));
                v.put(COL_UPDATED, System.currentTimeMillis());
                db.insertWithOnConflict(TABLE, null, v, SQLiteDatabase.CONFLICT_REPLACE);
            }
            db.setTransactionSuccessful();
        } catch (Exception ignored) {
        } finally {
            db.endTransaction();
        }
    }

    @JavascriptInterface
    public void clearAll() {
        w().delete(TABLE, null, null);
    }
}