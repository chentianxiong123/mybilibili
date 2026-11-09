package com.mybilibili.common.service;

public class VideoProcessSupport {

    public static class ProcessProgress {
        private int progress;
        private String stage;
        private String status;
        private boolean hasSubtitle;
        private boolean hasSummary;

        public int getProgress() { return progress; }
        public void setProgress(int progress) { this.progress = progress; }
        public String getStage() { return stage; }
        public void setStage(String stage) { this.stage = stage; }
        public String getStatus() { return status; }
        public void setStatus(String status) { this.status = status; }
        public boolean isHasSubtitle() { return hasSubtitle; }
        public void setHasSubtitle(boolean hasSubtitle) { this.hasSubtitle = hasSubtitle; }
        public boolean isHasSummary() { return hasSummary; }
        public void setHasSummary(boolean hasSummary) { this.hasSummary = hasSummary; }
    }
}
