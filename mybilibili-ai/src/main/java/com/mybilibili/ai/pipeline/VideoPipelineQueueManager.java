package com.mybilibili.ai.pipeline;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

import javax.annotation.PreDestroy;
import java.time.LocalDateTime;
import java.util.*;
import java.util.concurrent.*;
import java.util.concurrent.atomic.AtomicInteger;

/**
 * 全流程任务队列管理器
 * 确保单线程顺序执行，并发数为1
 */
@Component
public class VideoPipelineQueueManager {

    private static final Logger log = LoggerFactory.getLogger(VideoPipelineQueueManager.class);

    /**
     * 任务队列（阻塞队列，支持优先级）
     */
    private final PriorityBlockingQueue<PipelineTask> taskQueue = new PriorityBlockingQueue<>(
            100,
            Comparator.comparing(task -> task.getCreateTime())
    );

    /**
     * 正在处理的任务
     */
    private final Map<String, PipelineTask> processingTasks = new ConcurrentHashMap<>();

    /**
     * 已完成的任务（保留最近100个）
     */
    private final LinkedHashMap<String, PipelineTask> completedTasks = new LinkedHashMap<String, PipelineTask>(100, 0.75f, true) {
        @Override
        protected boolean removeEldestEntry(Map.Entry<String, PipelineTask> eldest) {
            return size() > 100;
        }
    };

    /**
     * 单线程执行器（确保并发数为1）
     */
    private final ExecutorService executor = Executors.newSingleThreadExecutor(r -> {
        Thread thread = new Thread(r, "video-pipeline-processor");
        thread.setDaemon(true);
        return thread;
    });

    /**
     * 任务处理器
     */
    private PipelineTaskProcessor taskProcessor;

    /**
     * 运行状态
     */
    private volatile boolean running = true;

    /**
     * 当前正在处理的任务
     */
    private volatile PipelineTask currentTask = null;

    /**
     * 统计信息
     */
    private final AtomicInteger totalSubmitted = new AtomicInteger(0);
    private final AtomicInteger totalCompleted = new AtomicInteger(0);
    private final AtomicInteger totalFailed = new AtomicInteger(0);

    /**
     * 设置任务处理器
     */
    public void setTaskProcessor(PipelineTaskProcessor processor) {
        this.taskProcessor = processor;
    }

    /**
     * 初始化并启动队列处理器
     */
    public void start() {
        log.info("[队列管理器] 启动任务处理器");
        executor.submit(this::processQueue);
    }

    /**
     * 提交任务到队列
     */
    public boolean submitTask(PipelineTask task) {
        if (!running) {
            log.warn("[队列管理器] 管理器已停止，拒绝任务: {}", task.getTaskKey());
            return false;
        }

        String taskKey = task.getTaskKey();

        // 检查是否已在队列中或正在处理
        if (isTaskInQueue(taskKey) || processingTasks.containsKey(taskKey)) {
            log.warn("[队列管理器] 任务已存在: {}", taskKey);
            return false;
        }

        taskQueue.offer(task);
        totalSubmitted.incrementAndGet();
        log.info("[队列管理器] 任务已加入队列: {}, 当前队列大小: {}", taskKey, taskQueue.size());
        return true;
    }

    /**
     * 取消任务
     */
    public boolean cancelTask(String taskKey) {
        // 尝试从队列中移除
        Optional<PipelineTask> taskInQueue = taskQueue.stream()
                .filter(t -> t.getTaskKey().equals(taskKey))
                .findFirst();

        if (taskInQueue.isPresent()) {
            taskQueue.remove(taskInQueue.get());
            taskInQueue.get().markCancelled();
            log.info("[队列管理器] 任务已从队列中取消: {}", taskKey);
            return true;
        }

        // 检查是否正在处理
        PipelineTask processing = processingTasks.get(taskKey);
        if (processing != null) {
            log.warn("[队列管理器] 任务正在处理中，无法取消: {}", taskKey);
            return false;
        }

        return false;
    }

    /**
     * 队列处理循环
     */
    private void processQueue() {
        log.info("[队列管理器] 队列处理线程启动");

        while (running) {
            try {
                // 从队列中获取任务（阻塞等待）
                PipelineTask task = taskQueue.poll(1, TimeUnit.SECONDS);

                if (task == null) {
                    continue;
                }

                String taskKey = task.getTaskKey();
                log.info("[队列管理器] 开始处理任务: {}", taskKey);

                // 标记为处理中
                currentTask = task;
                processingTasks.put(taskKey, task);
                task.markStarted();

                try {
                    // 执行任务
                    if (taskProcessor != null) {
                        taskProcessor.process(task);
                    }

                    // 检查任务状态
                    if (task.getStatus() == PipelineTask.TaskStatus.COMPLETED) {
                        totalCompleted.incrementAndGet();
                        log.info("[队列管理器] 任务完成: {}", taskKey);
                    } else if (task.getStatus() == PipelineTask.TaskStatus.FAILED) {
                        totalFailed.incrementAndGet();
                        log.error("[队列管理器] 任务失败: {}, 步骤: {}, 错误: {}",
                                taskKey, task.getFailedStep(), task.getErrorMessage());
                    }

                } catch (Exception e) {
                    log.error("[队列管理器] 任务处理异常: {}", taskKey, e);
                    task.markFailed("UNKNOWN", e.getMessage());
                    totalFailed.incrementAndGet();
                } finally {
                    // 移动到已完成列表
                    processingTasks.remove(taskKey);
                    synchronized (completedTasks) {
                        completedTasks.put(taskKey, task);
                    }
                    currentTask = null;
                }

            } catch (InterruptedException e) {
                log.info("[队列管理器] 队列处理线程被中断");
                Thread.currentThread().interrupt();
                break;
            } catch (Exception e) {
                log.error("[队列管理器] 队列处理异常", e);
            }
        }

        log.info("[队列管理器] 队列处理线程停止");
    }

    /**
     * 检查任务是否在队列中
     */
    public boolean isTaskInQueue(String taskKey) {
        return taskQueue.stream().anyMatch(t -> t.getTaskKey().equals(taskKey));
    }

    /**
     * 获取任务状态
     */
    public PipelineTask getTaskStatus(String taskKey) {
        // 检查正在处理的任务
        PipelineTask processing = processingTasks.get(taskKey);
        if (processing != null) {
            return processing;
        }

        // 检查队列中的任务
        Optional<PipelineTask> inQueue = taskQueue.stream()
                .filter(t -> t.getTaskKey().equals(taskKey))
                .findFirst();
        if (inQueue.isPresent()) {
            return inQueue.get();
        }

        // 检查已完成的任务
        synchronized (completedTasks) {
            return completedTasks.get(taskKey);
        }
    }

    /**
     * 获取队列大小
     */
    public int getQueueSize() {
        return taskQueue.size();
    }

    /**
     * 获取当前正在处理的任务
     */
    public PipelineTask getCurrentTask() {
        return currentTask;
    }

    /**
     * 获取统计信息
     */
    public Map<String, Object> getStatistics() {
        Map<String, Object> stats = new HashMap<>();
        stats.put("running", running);
        stats.put("queueSize", taskQueue.size());
        stats.put("processingCount", processingTasks.size());
        stats.put("totalSubmitted", totalSubmitted.get());
        stats.put("totalCompleted", totalCompleted.get());
        stats.put("totalFailed", totalFailed.get());
        stats.put("currentTask", currentTask != null ? currentTask.toString() : null);
        return stats;
    }

    /**
     * 获取队列中的所有任务
     */
    public List<PipelineTask> getQueuedTasks() {
        return new ArrayList<>(taskQueue);
    }

    /**
     * 获取正在处理的任务
     */
    public List<PipelineTask> getProcessingTasks() {
        return new ArrayList<>(processingTasks.values());
    }

    /**
     * 获取已完成的任务
     */
    public List<PipelineTask> getCompletedTasks() {
        synchronized (completedTasks) {
            return new ArrayList<>(completedTasks.values());
        }
    }

    /**
     * 清空队列
     */
    public void clearQueue() {
        taskQueue.clear();
        log.info("[队列管理器] 队列已清空");
    }

    /**
     * 停止管理器
     */
    @PreDestroy
    public void shutdown() {
        log.info("[队列管理器] 开始停止...");
        running = false;
        executor.shutdown();
        try {
            if (!executor.awaitTermination(30, TimeUnit.SECONDS)) {
                executor.shutdownNow();
            }
        } catch (InterruptedException e) {
            executor.shutdownNow();
            Thread.currentThread().interrupt();
        }
        log.info("[队列管理器] 已停止");
    }

    /**
     * 任务处理器接口
     */
    @FunctionalInterface
    public interface PipelineTaskProcessor {
        void process(PipelineTask task);
    }
}
