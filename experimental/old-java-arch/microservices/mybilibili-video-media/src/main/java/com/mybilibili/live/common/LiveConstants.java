package com.mybilibili.live.common;

import java.util.Set;

public final class LiveConstants {

    private LiveConstants() {
    }

    public static final class RoomStatus {
        public static final String LIVE = "live";
        public static final String OFFLINE = "offline";

        private static final Set<String> VALUES = Set.of(LIVE, OFFLINE);

        private RoomStatus() {
        }

        public static boolean isValid(String status) {
            return VALUES.contains(status);
        }
    }

    public static final class SrsAction {
        public static final String ON_PUBLISH = "on_publish";
        public static final String ON_UNPUBLISH = "on_unpublish";

        private SrsAction() {
        }

        public static String roomStatusFor(String action) {
            if (ON_PUBLISH.equals(action)) {
                return RoomStatus.LIVE;
            }
            if (ON_UNPUBLISH.equals(action)) {
                return RoomStatus.OFFLINE;
            }
            return null;
        }
    }

    public static final class MeetingStatus {
        public static final int NOT_STARTED = 0;
        public static final int IN_PROGRESS = 1;
        public static final int ENDED = 2;
        public static final int PENDING_APPROVAL = 3;
        public static final int REJECTED = 4;

        private static final Set<Integer> VALUES = Set.of(
                NOT_STARTED,
                IN_PROGRESS,
                ENDED,
                PENDING_APPROVAL,
                REJECTED
        );

        private MeetingStatus() {
        }

        public static boolean isValid(Integer status) {
            return status != null && VALUES.contains(status);
        }

        public static boolean isNotStarted(Integer status) {
            return Integer.valueOf(NOT_STARTED).equals(status);
        }

        public static boolean isPendingApproval(Integer status) {
            return Integer.valueOf(PENDING_APPROVAL).equals(status);
        }

        public static boolean isRejected(Integer status) {
            return Integer.valueOf(REJECTED).equals(status);
        }

        public static boolean isEnded(Integer status) {
            return Integer.valueOf(ENDED).equals(status);
        }

        public static boolean isJoinable(Integer status) {
            return Integer.valueOf(NOT_STARTED).equals(status)
                    || Integer.valueOf(IN_PROGRESS).equals(status);
        }
    }

    public static final class ParticipantRole {
        public static final int PARTICIPANT = 0;
        public static final int HOST = 1;

        private ParticipantRole() {
        }
    }

    public static final class SwitchState {
        public static final int OFF = 0;
        public static final int ON = 1;

        private SwitchState() {
        }
    }

    public static final class LinkmicStatus {
        public static final int PENDING = 0;
        public static final int CONNECTED = 1;
        public static final int DISCONNECTED = 2;
        public static final int REJECTED = 3;

        private LinkmicStatus() {
        }

        public static boolean isPending(Integer status) {
            return Integer.valueOf(PENDING).equals(status);
        }

        public static boolean isConnected(Integer status) {
            return Integer.valueOf(CONNECTED).equals(status);
        }

        public static boolean isFinished(Integer status) {
            return Integer.valueOf(DISCONNECTED).equals(status)
                    || Integer.valueOf(REJECTED).equals(status);
        }
    }
}
