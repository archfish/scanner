// 基础工具类 - 通用工具函数
class Utils {
    static handleFetchError(error, operation) {
        UIManager.showError(`${operation}时发生错误: ${error.message}`);
    }

    static confirmAction(message) {
        return confirm(message);
    }

    static processFetchResponse(response) {
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        return response.json();
    }

    static debounce(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    }
}

// DOM元素管理器 - 单例模式
class DOMManager {
    constructor() {
        if (DOMManager.instance) {
            return DOMManager.instance;
        }
        this.elements = {};
        this.initializeElements();
        DOMManager.instance = this;
    }

    initializeElements() {
        const elementIds = [
            'deviceList', 'scanForm', 'scanBtn', 'scanBtnText', 'scanProgress',
            'scanStatus', 'previewPlaceholder', 'imageContainer', 'imageInfo',
            'imageDimensions', 'imageSize', 'imageCanvas', 'zoomInBtn', 'zoomOutBtn',
            'fitBtn', 'downloadBtn', 'scanHistory', 'historyPlaceholder', 'toast', 'toastMessage'
        ];

        elementIds.forEach(id => {
            this.elements[id] = document.getElementById(id);
        });
    }

    get(elementId) {
        return this.elements[elementId];
    }
}

// 状态管理器 - 单例模式
class StateManager {
    constructor() {
        if (StateManager.instance) {
            return StateManager.instance;
        }
        this.state = {
            selectedDevice: null,
            scanHistory: [],
            canvasState: {
                scale: 1,
                offsetX: 0,
                offsetY: 0,
                isDragging: false,
                startX: 0,
                startY: 0,
                lastX: 0,
                lastY: 0,
                currentImage: null
            }
        };
        StateManager.instance = this;
    }

    updateSelectedDevice(device) {
        this.state.selectedDevice = device;
    }

    getSelectedDevice() {
        return this.state.selectedDevice;
    }

    updateCanvasState(updates) {
        Object.assign(this.state.canvasState, updates);
    }

    getCanvasState() {
        return this.state.canvasState;
    }

    resetCanvasState() {
        this.state.canvasState = {
            scale: 1,
            offsetX: 0,
            offsetY: 0,
            isDragging: false,
            startX: 0,
            startY: 0,
            lastX: 0,
            lastY: 0,
            currentImage: null
        };
    }

    addToHistory(record) {
        this.state.scanHistory.unshift(record);
        if (this.state.scanHistory.length > 10) {
            this.state.scanHistory.pop();
        }
    }

    getHistory() {
        return this.state.scanHistory;
    }

    clearHistory() {
        this.state.scanHistory = [];
    }

    setHistory(history) {
        this.state.scanHistory = history;
    }
}

// 设备管理器 - 策略模式
class DeviceManager {
    static loadDevices() {
        const dom = new DOMManager();
        UIManager.showStatus('正在加载设备...');

        return fetch('/api/devices')
            .then(Utils.processFetchResponse)
            .then(data => {
                return DeviceManager.handleDevicesResponse(data);
            })
            .catch(error => {
                Utils.handleFetchError(error, '加载设备');
            });
    }

    static handleDevicesResponse(data) {
        if (data.Code !== '0') {
            UIManager.showError('加载设备失败: ' + data.Msg);
            return;
        }

        DeviceManager.renderDeviceList(data.Data);
        UIManager.showStatus('设备加载完成');
        return data.Data;
    }

    static renderDeviceList(devices) {
        const dom = new DOMManager();
        const state = new StateManager();

        if (!devices || devices.length === 0) {
            dom.get('deviceList').innerHTML = DeviceManager.getEmptyDeviceTemplate();
            return;
        }

        const html = devices.map((device, index) =>
            DeviceManager.getDeviceItemTemplate(device, index)
        ).join('');

        dom.get('deviceList').innerHTML = html;

        // 默认选择第一个设备
        if (devices.length > 0) {
            state.updateSelectedDevice(devices[0]);
        }

        DeviceManager.bindDeviceEvents(devices);
    }

    static getEmptyDeviceTemplate() {
        return `
                    <div class="text-center py-5">
                        <div class="placeholder-icon" style="font-size: 2rem;">🔌</div>
                        <p class="mb-1">未找到可用设备</p>
                        <p class="text-muted small">请检查设备连接</p>
                    </div>
                `;
    }

    static getDeviceItemTemplate(device, index) {
        const activeClass = index === 0 ? 'active' : '';
        return `
                    <div class="device-item ${activeClass}" data-index="${index}">
                        <div class="device-name">${device.Name || '未知设备'}</div>
                        <div class="device-id">${device.VendorID}:${device.ProductID}</div>
                    </div>
                `;
    }

    static bindDeviceEvents(devices) {
        const state = new StateManager();
        document.querySelectorAll('.device-item').forEach(item => {
            item.addEventListener('click', function () {
                document.querySelectorAll('.device-item').forEach(i => i.classList.remove('active'));
                this.classList.add('active');
                const index = parseInt(this.getAttribute('data-index'));
                state.updateSelectedDevice(devices[index]);
            });
        });
    }
}

// 扫描管理器 - 命令模式
class ScanManager {
    static handleScan(event) {
        event.preventDefault();

        const state = new StateManager();
        const selectedDevice = state.getSelectedDevice();

        if (!selectedDevice) {
            UIManager.showError('请先选择一个设备');
            return;
        }

        const scanOptions = ScanManager.getScanOptions();
        const requestData = {
            device: selectedDevice,
            option: scanOptions
        };

        ScanManager.executeScan(requestData, scanOptions);
    }

    static getScanOptions() {
        return {
            DPI: parseInt(document.getElementById('dpi').value),
            Mode: document.getElementById('mode').value,
            Width: parseFloat(document.getElementById('width').value),
            Height: parseFloat(document.getElementById('height').value),
            Left: parseFloat(document.getElementById('left').value),
            Top: parseFloat(document.getElementById('top').value)
        };
    }

    static executeScan(requestData, scanOptions) {
        UIManager.disableScanButton();
        const progressController = new ProgressController();
        progressController.start();

        fetch('/api/scan', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(requestData)
        })
            .then(Utils.processFetchResponse)
            .then(data => {
                progressController.complete();
                return ScanManager.handleScanResponse(data, requestData, scanOptions);
            })
            .catch(error => {
                progressController.error();
                Utils.handleFetchError(error, '扫描');
                UIManager.resetScanButton();
            });
    }

    static handleScanResponse(data, requestData, scanOptions) {
        if (data.Code !== '0') {
            UIManager.showError('扫描失败: ' + data.Msg);
            UIManager.resetScanButton();
            return;
        }

        UIManager.showStatus('扫描完成');
        ImageManager.displayImage(data.Data.URL);
        HistoryManager.addToScanHistory({
            device: requestData.device,
            options: scanOptions,
            filePath: data.Data.URL,
            timestamp: new Date().toLocaleString()
        });
        UIManager.resetScanButton();
    }
}

// 进度控制器
class ProgressController {
    constructor() {
        this.progress = 0;
        this.interval = null;
        this.dom = new DOMManager();
    }

    start() {
        this.progress = 0;
        this.dom.get('scanProgress').style.width = '0%';
        UIManager.showStatus('正在准备扫描...');

        this.interval = setInterval(() => {
            this.progress += 95 / 350; // 35秒达到95%
            if (this.progress >= 95) {
                this.progress = 95;
            }
            this.dom.get('scanProgress').style.width = this.progress + '%';
            UIManager.showStatus(`扫描中... ${Math.round(this.progress)}%`);
        }, 100);
    }

    complete() {
        if (this.interval) {
            clearInterval(this.interval);
            this.interval = null;
        }
        this.dom.get('scanProgress').style.width = '100%';
    }

    error() {
        if (this.interval) {
            clearInterval(this.interval);
            this.interval = null;
        }
        this.dom.get('scanProgress').style.width = '100%';
    }
}

// 图像管理器 - 观察者模式
class ImageManager {
    static displayImage(filePath) {
        const img = new Image();
        img.crossOrigin = 'anonymous';
        img.src = filePath;

        img.onload = () => {
            ImageManager.handleImageLoad(img, filePath);
        };

        img.onerror = () => {
            UIManager.showError('图片加载失败，请检查文件路径');
        };
    }

    static handleImageLoad(img, filePath) {
        const dom = new DOMManager();
        const state = new StateManager();

        dom.get('imageContainer').style.display = 'block';
        dom.get('previewPlaceholder').style.display = 'none';

        setTimeout(() => {
            ImageManager.setupCanvas(img);
            ImageManager.updateImageInfo(img);

            // 存储当前图像
            state.updateCanvasState({ currentImage: img });

            const canvasController = new CanvasController(img);
            canvasController.initialize();

            UIManager.enableViewerButtons();
            UIManager.setupDownloadButton(filePath);
            UIManager.resetScanButton();
        }, 100);
    }

    static setupCanvas(img) {
        const dom = new DOMManager();
        const container = dom.get('imageContainer');

        // 确保容器有明确的尺寸
        const containerRect = container.getBoundingClientRect();
        const containerWidth = Math.max(containerRect.width, 600);
        const containerHeight = Math.max(containerRect.height, 400);

        const canvas = dom.get('imageCanvas');
        canvas.width = containerWidth;
        canvas.height = containerHeight;
        canvas.style.display = 'block';
        canvas.style.width = containerWidth + 'px';
        canvas.style.height = containerHeight + 'px';
    }

    static updateImageInfo(img) {
        const dom = new DOMManager();
        dom.get('imageDimensions').textContent = `尺寸: ${img.width}×${img.height}px`;
        dom.get('imageSize').textContent = '大小: -';
        dom.get('imageInfo').style.display = 'block';
    }
}

// 画布控制器 - 责任链模式
class CanvasController {
    constructor(img) {
        this.img = img;
        this.dom = new DOMManager();
        this.state = new StateManager();
        this.canvas = this.dom.get('imageCanvas');
        this.ctx = this.canvas.getContext('2d');

        // 事件处理器链
        this.eventHandlers = {
            wheel: this.handleWheel.bind(this),
            mousedown: this.handleMouseDown.bind(this),
            mousemove: this.handleMouseMove.bind(this),
            mouseup: this.handleMouseUp.bind(this),
            mouseleave: this.handleMouseUp.bind(this),
            touchstart: this.handleTouchStart.bind(this),
            touchmove: this.handleTouchMove.bind(this),
            touchend: this.handleTouchEnd.bind(this)
        };
    }

    initialize() {
        this.state.resetCanvasState();
        this.removeAllEventListeners();
        this.addEventListeners();
        this.setupButtonEvents();
        this.fitToView();
    }

    removeAllEventListeners() {
        Object.keys(this.eventHandlers).forEach(eventType => {
            this.canvas.removeEventListener(eventType, this.eventHandlers[eventType]);
        });
    }

    addEventListeners() {
        Object.entries(this.eventHandlers).forEach(([eventType, handler]) => {
            this.canvas.addEventListener(eventType, handler);
        });
    }

    setupButtonEvents() {
        this.dom.get('fitBtn').onclick = () => this.fitToView();
        this.dom.get('zoomInBtn').onclick = () => this.zoomIn();
        this.dom.get('zoomOutBtn').onclick = () => this.zoomOut();
    }

    // 修复的居中算法 - 确保图片完全适配容器
    calculateFitParams() {
        const canvasWidth = this.canvas.width;
        const canvasHeight = this.canvas.height;
        const imgWidth = this.img.width;
        const imgHeight = this.img.height;

        // 计算缩放比例，使图片完全适配容器
        const scaleX = canvasWidth / imgWidth;
        const scaleY = canvasHeight / imgHeight;
        const scale = Math.min(scaleX, scaleY);

        // 计算居中位置
        const scaledWidth = imgWidth * scale;
        const scaledHeight = imgHeight * scale;
        const x = (canvasWidth - scaledWidth) / 2;
        const y = (canvasHeight - scaledHeight) / 2;

        return { scale, x, y, scaledWidth, scaledHeight };
    }

    redraw() {
        const canvasState = this.state.getCanvasState();
        this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);
        this.ctx.save();

        // 应用变换 - 简化变换逻辑，直接以画布中心为基准
        this.ctx.translate(
            this.canvas.width / 2 + canvasState.offsetX,
            this.canvas.height / 2 + canvasState.offsetY
        );
        this.ctx.scale(canvasState.scale, canvasState.scale);

        // 居中绘制图片 - 以图片中心为原点
        this.ctx.drawImage(
            this.img,
            -this.img.width / 2,
            -this.img.height / 2,
            this.img.width,
            this.img.height
        );

        this.ctx.restore();
    }

    fitToView() {
        const fitParams = this.calculateFitParams();
        this.state.updateCanvasState({
            scale: fitParams.scale,
            offsetX: 0,
            offsetY: 0
        });
        this.redraw();
    }

    zoomIn() {
        const canvasState = this.state.getCanvasState();
        const newScale = Math.min(canvasState.scale * 1.2, 5);
        this.state.updateCanvasState({ scale: newScale });
        this.redraw();
    }

    zoomOut() {
        const canvasState = this.state.getCanvasState();
        const newScale = Math.max(canvasState.scale / 1.2, 0.1);
        this.state.updateCanvasState({ scale: newScale });
        this.redraw();
    }

    handleWheel(e) {
        e.preventDefault();

        const canvasState = this.state.getCanvasState();
        const rect = this.canvas.getBoundingClientRect();
        const mouseX = e.clientX - rect.left - this.canvas.width / 2;
        const mouseY = e.clientY - rect.top - this.canvas.height / 2;

        // 计算缩放前鼠标在图片中的位置
        const beforeX = (mouseX - canvasState.offsetX) / canvasState.scale;
        const beforeY = (mouseY - canvasState.offsetY) / canvasState.scale;

        // 应用缩放
        let newScale = canvasState.scale;
        if (e.deltaY < 0) {
            newScale *= 1.1;
        } else {
            newScale /= 1.1;
        }
        newScale = Math.max(0.1, Math.min(newScale, 5));

        // 计算新的偏移量，保持鼠标点不变
        const newOffsetX = mouseX - beforeX * newScale;
        const newOffsetY = mouseY - beforeY * newScale;

        this.state.updateCanvasState({
            scale: newScale,
            offsetX: newOffsetX,
            offsetY: newOffsetY
        });

        this.redraw();
    }

    handleMouseDown(e) {
        const canvasState = this.state.getCanvasState();
        this.state.updateCanvasState({
            isDragging: true,
            startX: e.clientX,
            startY: e.clientY,
            lastX: canvasState.offsetX,
            lastY: canvasState.offsetY
        });
        this.canvas.style.cursor = 'grabbing';
        e.preventDefault();
    }

    handleMouseMove(e) {
        const canvasState = this.state.getCanvasState();
        if (!canvasState.isDragging) return;

        e.preventDefault();
        const deltaX = e.clientX - canvasState.startX;
        const deltaY = e.clientY - canvasState.startY;

        this.state.updateCanvasState({
            offsetX: canvasState.lastX + deltaX,
            offsetY: canvasState.lastY + deltaY
        });

        this.redraw();
    }

    handleMouseUp(e) {
        this.state.updateCanvasState({ isDragging: false });
        this.canvas.style.cursor = 'move';
    }

    handleTouchStart(e) {
        if (e.touches.length !== 1) return;

        const canvasState = this.state.getCanvasState();
        this.state.updateCanvasState({
            isDragging: true,
            startX: e.touches[0].clientX,
            startY: e.touches[0].clientY,
            lastX: canvasState.offsetX,
            lastY: canvasState.offsetY
        });
        e.preventDefault();
    }

    handleTouchMove(e) {
        const canvasState = this.state.getCanvasState();
        if (!canvasState.isDragging || e.touches.length !== 1) return;

        e.preventDefault();
        const deltaX = e.touches[0].clientX - canvasState.startX;
        const deltaY = e.touches[0].clientY - canvasState.startY;

        this.state.updateCanvasState({
            offsetX: canvasState.lastX + deltaX,
            offsetY: canvasState.lastY + deltaY
        });

        this.redraw();
    }

    handleTouchEnd(e) {
        this.state.updateCanvasState({ isDragging: false });
    }
}

// 历史记录管理器 - 存储库模式
class HistoryManager {
    static addToScanHistory(record) {
        const state = new StateManager();
        state.addToHistory(record);

        localStorage.setItem('scanHistory', JSON.stringify(state.getHistory()));
        SettingsManager.saveSettings();
        HistoryManager.renderScanHistory();
    }

    static loadScanHistory() {
        const savedHistory = localStorage.getItem('scanHistory');
        if (!savedHistory) return;

        const state = new StateManager();
        state.setHistory(JSON.parse(savedHistory));
        HistoryManager.renderScanHistory();
    }

    static renderScanHistory() {
        const dom = new DOMManager();
        const state = new StateManager();
        const history = state.getHistory();

        if (history.length === 0) {
            dom.get('historyPlaceholder').style.display = 'block';
            dom.get('scanHistory').innerHTML = '';
            return;
        }

        dom.get('historyPlaceholder').style.display = 'none';
        const html = history.map((record, index) =>
            HistoryManager.getHistoryItemTemplate(record)
        ).join('');

        dom.get('scanHistory').innerHTML = html;
    }

    static getHistoryItemTemplate(record) {
        return `
                    <div class="history-item">
                        <div class="history-img">
                            <img src="${record.filePath}" alt="历史扫描结果" style="width:100%; height:150px; object-fit: cover;">
                        </div>
                        <div class="history-content">
                            <div class="history-title">${record.device.Name || '未知设备'}</div>
                            <div class="history-info">📅 ${record.timestamp}</div>
                            <div class="history-info">📐 ${record.options.Width}×${record.options.Height}mm</div>
                            <button class="btn btn-outline mt-2" style="padding: 5px 10px; font-size: 0.8rem;" onclick="HistoryManager.viewHistoryImage('${record.filePath}')">
                                👁️ 查看
                            </button>
                        </div>
                    </div>
                `;
    }

    static clearScanHistory() {
        if (!Utils.confirmAction('确定要清空所有扫描历史记录吗？此操作将删除服务器上的所有扫描文件。')) return;

        fetch('/api/attachments', {
            method: 'DELETE'
        })
            .then(Utils.processFetchResponse)
            .then(data => {
                return HistoryManager.handleClearResponse(data);
            })
            .catch(error => {
                Utils.handleFetchError(error, '清空历史记录');
            });
    }

    static handleClearResponse(data) {
        if (data.Code !== '0') {
            UIManager.showError('清空失败: ' + data.Msg);
            return;
        }

        const state = new StateManager();
        state.clearHistory();
        localStorage.removeItem('scanHistory');
        HistoryManager.renderScanHistory();
        UIManager.showSuccess('扫描历史已清空');
        HistoryManager.clearPreview();
    }

    static clearPreview() {
        const dom = new DOMManager();
        dom.get('imageContainer').style.display = 'none';
        dom.get('previewPlaceholder').style.display = 'block';
        dom.get('imageInfo').style.display = 'none';

        // 禁用按钮
        ['downloadBtn', 'zoomInBtn', 'zoomOutBtn', 'fitBtn'].forEach(btnId => {
            dom.get(btnId).disabled = true;
        });
    }

    static viewHistoryImage(filePath) {
        ImageManager.displayImage(filePath);
        UIManager.showSuccess('已加载历史图片');
    }
}

// 设置管理器 - 单例模式
class SettingsManager {
    static saveSettings() {
        const state = new StateManager();
        const selectedDevice = state.getSelectedDevice();

        if (!selectedDevice) return;

        const settings = {
            device: selectedDevice,
            options: SettingsManager.getCurrentOptions()
        };
        localStorage.setItem('scannerSettings', JSON.stringify(settings));
    }

    static getCurrentOptions() {
        return {
            dpi: document.getElementById('dpi').value,
            mode: document.getElementById('mode').value,
            width: document.getElementById('width').value,
            height: document.getElementById('height').value,
            left: document.getElementById('left').value,
            top: document.getElementById('top').value
        };
    }

    static loadSavedSettings() {
        const savedSettings = localStorage.getItem('scannerSettings');
        if (!savedSettings) return;

        try {
            const settings = JSON.parse(savedSettings);
            const state = new StateManager();

            if (settings.device) {
                state.updateSelectedDevice(settings.device);
            }

            if (settings.options) {
                SettingsManager.applyOptions(settings.options);
            }
        } catch (e) {
            console.error('Failed to load saved settings:', e);
        }
    }

    static applyOptions(options) {
        const optionMap = {
            dpi: options.dpi || '400',
            mode: options.mode || 'CGRAY',
            width: options.width || '211.881',
            height: options.height || '355.567',
            left: options.left || '0',
            top: options.top || '0'
        };

        Object.entries(optionMap).forEach(([id, value]) => {
            const element = document.getElementById(id);
            if (element) element.value = value;
        });
    }
}

// UI管理器 - 外观模式
class UIManager {
    static setupEventListeners() {
        const dom = new DOMManager();

        document.getElementById('refreshDevices').addEventListener('click',
            () => DeviceManager.loadDevices());
        dom.get('scanForm').addEventListener('submit',
            (event) => ScanManager.handleScan(event));
        document.getElementById('clearHistoryBtn').addEventListener('click',
            () => HistoryManager.clearScanHistory());
    }

    static showStatus(message) {
        const dom = new DOMManager();
        dom.get('scanStatus').textContent = message;
    }

    static showError(message) {
        UIManager.showStatus(message);
        UIManager.showToast(message, 'error');
    }

    static showSuccess(message) {
        UIManager.showToast(message, 'success');
    }

    static showToast(message, type = 'info') {
        const dom = new DOMManager();
        dom.get('toastMessage').textContent = message;
        dom.get('toast').className = 'toast show';

        // 使用策略模式处理不同类型的toast
        const toastStrategies = {
            error: () => dom.get('toast').classList.add('error'),
            success: () => dom.get('toast').classList.add('success'),
            info: () => { } // 默认样式
        };

        const strategy = toastStrategies[type] || toastStrategies.info;
        strategy();

        setTimeout(() => {
            dom.get('toast').classList.remove('show');
        }, 3000);
    }

    static disableScanButton() {
        const dom = new DOMManager();
        dom.get('scanBtn').disabled = true;
        dom.get('scanBtnText').innerHTML = '<span class="spinner" style="border-width: 2px; width: 16px; height: 16px; margin-right: 8px;"></span>扫描中...';
    }

    static resetScanButton() {
        const dom = new DOMManager();
        dom.get('scanBtn').disabled = false;
        dom.get('scanBtnText').innerHTML = '🔍 开始扫描';
    }

    static enableViewerButtons() {
        const dom = new DOMManager();
        ['downloadBtn', 'zoomInBtn', 'zoomOutBtn', 'fitBtn'].forEach(btnId => {
            dom.get(btnId).disabled = false;
        });
    }

    static setupDownloadButton(filePath) {
        const dom = new DOMManager();
        dom.get('downloadBtn').onclick = function () {
            const link = document.createElement('a');
            link.href = filePath;
            const urlParts = filePath.split('/');
            const filename = urlParts[urlParts.length - 1] || 'scan-result.jpg';
            link.download = filename;
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);
        };
    }
}

// 主应用控制器 - 门面模式
class ScannerApp {
    static init() {
        // 初始化各个管理器
        new DOMManager(); // 确保DOM管理器被初始化
        new StateManager(); // 确保状态管理器被初始化

        DeviceManager.loadDevices();
        UIManager.setupEventListeners();
        HistoryManager.loadScanHistory();
        SettingsManager.loadSavedSettings();
    }
}

// DOM 加载完成后初始化应用
document.addEventListener('DOMContentLoaded', function () {
    ScannerApp.init();
});