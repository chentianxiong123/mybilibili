package com.mybilibili.ai.process;

public class VideoProcessContext {

    private Integer videoId;
    private Integer manuscriptId;
    private Integer userId;
    private String videoTitle;
    private String videoDescription;
    private VideoProcessStepType currentStep;
    private ProcessMode processMode;
    private String triggerSource;

    public Integer getVideoId() {
        return videoId;
    }

    public void setVideoId(Integer videoId) {
        this.videoId = videoId;
    }

    public Integer getManuscriptId() {
        return manuscriptId;
    }

    public void setManuscriptId(Integer manuscriptId) {
        this.manuscriptId = manuscriptId;
    }

    public Integer getUserId() {
        return userId;
    }

    public void setUserId(Integer userId) {
        this.userId = userId;
    }

    public String getVideoTitle() {
        return videoTitle;
    }

    public void setVideoTitle(String videoTitle) {
        this.videoTitle = videoTitle;
    }

    public String getVideoDescription() {
        return videoDescription;
    }

    public void setVideoDescription(String videoDescription) {
        this.videoDescription = videoDescription;
    }

    public VideoProcessStepType getCurrentStep() {
        return currentStep;
    }

    public void setCurrentStep(VideoProcessStepType currentStep) {
        this.currentStep = currentStep;
    }

    public ProcessMode getProcessMode() {
        return processMode;
    }

    public void setProcessMode(ProcessMode processMode) {
        this.processMode = processMode;
    }

    public String getTriggerSource() {
        return triggerSource;
    }

    public void setTriggerSource(String triggerSource) {
        this.triggerSource = triggerSource;
    }
}
