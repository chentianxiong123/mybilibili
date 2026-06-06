package com.mybilibili.live.websocket;

import java.util.Set;

final class MeetingWsMessageType {

    static final String PING = "ping";
    static final String PONG = "pong";
    static final String JOIN = "join";
    static final String LEAVE = "leave";
    static final String JOIN_REQUEST = "join-request";
    static final String ADMIT_USER = "admit-user";
    static final String REJECT_USER = "reject-user";
    static final String ROOM_USERS = "room-users";
    static final String USER_JOINED = "user-joined";
    static final String USER_LEFT = "user-left";
    static final String VIEWER_COUNT = "viewer-count";

    static final String OFFER = "offer";
    static final String ANSWER = "answer";
    static final String ICE_CANDIDATE = "ice-candidate";
    static final String LINKMIC_APPLY = "linkmic-apply";
    static final String LINKMIC_ACCEPTED = "linkmic-accepted";
    static final String LINKMIC_REJECTED = "linkmic-rejected";
    static final String LINKMIC_DISCONNECTED = "linkmic-disconnected";
    static final String TOGGLE_AUDIO = "toggle-audio";
    static final String TOGGLE_VIDEO = "toggle-video";
    static final String TOGGLE_SCREEN = "toggle-screen";
    static final String PEER_STATE = "peer-state";
    static final String KICK = "kick";
    static final String MUTE_TARGET = "mute-target";
    static final String MUTE_VIDEO_TARGET = "mute-video-target";
    static final String MUTE_ALL = "mute-all";
    static final String CHAT = "chat";
    static final String MEETING_ENDED = "meeting-ended";
    static final String HAND_RAISED = "hand-raised";
    static final String SPOTLIGHT = "spotlight";
    static final String LOCK_ROOM = "lock-room";
    static final String TRANSFER_HOST = "transfer-host";
    static final String REACTION = "reaction";
    static final String GIFT = "gift";
    static final String PIN_MESSAGE = "pin-message";
    static final String UNPIN_MESSAGE = "unpin-message";

    private static final Set<String> FORWARDED_TYPES = Set.of(
            OFFER,
            ANSWER,
            ICE_CANDIDATE,
            LINKMIC_APPLY,
            LINKMIC_ACCEPTED,
            LINKMIC_REJECTED,
            LINKMIC_DISCONNECTED,
            TOGGLE_AUDIO,
            TOGGLE_VIDEO,
            TOGGLE_SCREEN,
            PEER_STATE,
            KICK,
            MUTE_TARGET,
            MUTE_VIDEO_TARGET,
            MUTE_ALL,
            CHAT,
            MEETING_ENDED,
            HAND_RAISED,
            SPOTLIGHT,
            LOCK_ROOM,
            TRANSFER_HOST,
            REACTION,
            GIFT,
            PIN_MESSAGE,
            UNPIN_MESSAGE
    );

    private MeetingWsMessageType() {
    }

    static boolean isForwarded(String type) {
        return type != null && FORWARDED_TYPES.contains(type);
    }
}
