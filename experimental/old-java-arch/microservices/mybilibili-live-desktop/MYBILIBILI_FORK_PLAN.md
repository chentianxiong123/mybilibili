# mybilibili Live Desktop Fork Plan

This project is a GPL-3.0 fork of Streamlabs Desktop.

## Goal

Keep the Electron + OBS core and replace Streamlabs-specific product services with the mybilibili live workflow.

## Keep First

- `main.js`: Electron bootstrap, worker window, crash/update hooks.
- `obs-api/`: access wrapper for `obs-studio-node`.
- `app/app.ts`: OBS API initialization and service bootstrap.
- `app/services/scenes/`: scene graph and scene item management.
- `app/services/sources/`: OBS source management.
- `app/services/streaming/`: start/stop streaming and stream state.
- `app/services/settings/`: stream/video/audio configuration.
- `app/services/audio/`: mixer and audio device handling.
- `app/services/video-encoding-optimizations/`: encoder optimization.
- `app/components/obs/`: OBS properties UI controls.
- `electron-builder/`: packaging baseline.

## Replace

- Product name and window titles: Streamlabs Desktop -> mybilibili Live Desktop.
- Protocol scheme: `slobs://` -> `mybilibili-live://`.
- User data directory: `slobs-client` -> `mybilibili-live-desktop`.
- Platform services: Twitch/Youtube/Tiktok/Facebook/etc. -> mybilibili live room service.
- Login/auth: Streamlabs auth -> mybilibili auth.
- Stream metadata: platform-specific title/category/tags -> mybilibili room info.

## Cut Candidates

- `app/services/grow/`
- `app/services/platform-app-store/`
- `app/services/platform-apps/`
- `app/services/guest-cam/`
- `app/services/game-overlay/`
- `app/services/stream-avatar/`
- `app/services/widgets/` except pieces needed for overlays.
- `app/services/streamlabels/`
- Third-party platform services under `app/services/platforms/` except a new mybilibili service.
- Streamlabs cloud sync, store, themes, donation, and marketing flows.

## First Milestone

Status: implemented as the current integration baseline.

1. Boot the fork with original Streamlabs core still intact.
2. Remove or hide nonessential navigation entries.
3. Add a mybilibili live service that calls `/api/live/room/my`.
4. Auto-create a room with `/api/live/room/create` when needed.
5. Write OBS stream settings as `rtmp_custom` using `rtmp://{gateway-host}/live` and `LiveRoom.streamKey`.
6. Add a mybilibili live dashboard page as the default app page.
7. Keep OBS scene/source/streaming operations available through the original `Studio` page.
8. Skip Streamlabs onboarding and login prompts for the fork.

## Current Desktop Flow

1. Open the app into `MybilibiliLive`.
2. Enter the Gateway base URL, defaulting to `http://localhost:8080/api`.
3. Paste a mybilibili JWT token.
4. Load or create the user's live room.
5. Edit room name, category, and cover URL.
6. Apply OBS stream settings.
7. Use `Edit Scenes` to configure OBS scenes and sources.
8. Start RTMP streaming from the live control room.
9. SRS hook updates the room status to `live`; stop streaming marks it `offline`.

## Backend Contract

The first desktop integration uses existing `mybilibili-live` endpoints through Gateway:

- `GET /api/live/room/my`
- `POST /api/live/room/create`
- `PUT /api/live/room/{id}`
- `PUT /api/live/room/{id}/schedule`
- `PUT /api/live/room/{id}/status`

The desktop app does not need a new Java service for the first milestone. SRS already calls
`POST /api/live/room/srs/hook` to update room status from `offline` to `live` and back.

## Non-Goals For First Milestone

- Do not rewrite OBS integration.
- Do not embed a new OBS runtime strategy.
- Do not migrate Vue 2 to Vue 3 yet.
- Do not touch existing mybilibili Java services.
- Do not merge this into `mybilibili-web`.

## Next Milestones

- Replace manual JWT input with mybilibili desktop login/token handoff.
- Add linkmic controls only if they map cleanly to `mybilibili-live`; do not reuse Streamlabs Guest Cam.
- Remove or stub remaining Streamlabs cloud services after their imports are proven unused.
- Rename remaining visible product strings in title bars, dialogs, and settings.

## Trim Pass 2

Status: implemented as the current visible-product cleanup baseline.

1. Keep the independent desktop client; do not merge it into `mybilibili-web`.
2. Side navigation is reduced to the live console, scene editor, recording history, and settings.
3. Source showcase is reduced to OBS/local sources only: screen, window, game, camera, audio, image, media, text, scene.
4. Widget, platform app, app store, Guest Cam, Collab Cam, Reactive Source, Streamlabs ID, Prime/Ultra, multistream, and Stream Shift UI paths are hidden or replaced with blank fallbacks.
5. Stream settings are reduced to the OBS custom RTMP settings page.
6. General, appearance, and scene collection settings are localized and stripped of widget/store/platform growth entries.
