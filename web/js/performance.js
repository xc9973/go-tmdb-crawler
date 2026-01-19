/**
 * ========================================
 * Performance Optimization Toolkit
 * TMDB Crawler - Glassmorphism Upgrade
 * ========================================
 *
 * Features:
 * - Image lazy loading with Intersection Observer
 * - Debounce and throttle utilities
 * - Virtual scrolling for long lists
 * - Performance monitoring
 * - RequestAnimationFrame utilities
 * - Memory leak prevention
 */

(function(global) {
    'use strict';

    // ========================================
    // 1. Image Lazy Loading
    // ========================================

    const LazyLoader = {
        observer: null,
        loadedImages: new WeakSet(),

        /**
         * Initialize lazy loading for images
         * @param {string} selector - Image selector (default: 'img[data-lazy]')
         * @param {Object} options - Intersection Observer options
         */
        init(selector = 'img[data-lazy]', options = {}) {
            // Check if Intersection Observer is supported
            if (!('IntersectionObserver' in global)) {
                console.warn('IntersectionObserver not supported, loading all images immediately');
                this.loadAll(selector);
                return;
            }

            const defaultOptions = {
                rootMargin: '50px 0px', // Load 50px before entering viewport
                threshold: 0.01
            };

            this.observer = new IntersectionObserver((entries) => {
                entries.forEach(entry => {
                    if (entry.isIntersecting) {
                        this.loadImage(entry.target);
                        this.observer.unobserve(entry.target);
                    }
                });
            }, { ...defaultOptions, ...options });

            // Observe all lazy images
            document.querySelectorAll(selector).forEach(img => {
                if (!this.loadedImages.has(img)) {
                    this.observer.observe(img);
                }
            });
        },

        /**
         * Load a single image
         * @param {HTMLImageElement} img - Image element
         */
        loadImage(img) {
            if (this.loadedImages.has(img)) return;

            const src = img.dataset.lazy || img.dataset.src;
            if (!src) return;

            // Add loading class
            img.classList.add('lazy-loading');

            // Create new image to preload
            const tempImg = new Image();

            tempImg.onload = () => {
                img.src = src;
                img.classList.remove('lazy-loading');
                img.classList.add('lazy-loaded');
                this.loadedImages.add(img);

                // Remove attributes after loading
                delete img.dataset.lazy;
                delete img.dataset.src;
            };

            tempImg.onerror = () => {
                img.classList.remove('lazy-loading');
                img.classList.add('lazy-error');
                console.warn(`Failed to load image: ${src}`);
            };

            tempImg.src = src;
        },

        /**
         * Load all images immediately (fallback)
         * @param {string} selector - Image selector
         */
        loadAll(selector) {
            document.querySelectorAll(selector).forEach(img => {
                this.loadImage(img);
            });
        },

        /**
         * Disconnect observer and clean up
         */
        destroy() {
            if (this.observer) {
                this.observer.disconnect();
                this.observer = null;
            }
        }
    };

    // ========================================
    // 2. Debounce & Throttle Utilities
    // ========================================

    const PerformanceUtils = {
        /**
         * Debounce function - delays execution until after wait time has elapsed
         * @param {Function} func - Function to debounce
         * @param {number} wait - Wait time in ms
         * @param {boolean} immediate - Execute on leading edge
         * @returns {Function} Debounced function
         */
        debounce(func, wait = 150, immediate = false) {
            let timeout;
            return function executedFunction(...args) {
                const later = () => {
                    timeout = null;
                    if (!immediate) func.apply(this, args);
                };
                const callNow = immediate && !timeout;
                clearTimeout(timeout);
                timeout = setTimeout(later, wait);
                if (callNow) func.apply(this, args);
            };
        },

        /**
         * Throttle function - limits execution rate
         * @param {Function} func - Function to throttle
         * @param {number} limit - Minimum time between calls in ms
         * @returns {Function} Throttled function
         */
        throttle(func, limit = 100) {
            let inThrottle;
            return function(...args) {
                if (!inThrottle) {
                    func.apply(this, args);
                    inThrottle = true;
                    setTimeout(() => inThrottle = false, limit);
                }
            };
        },

        /**
         * RequestAnimationFrame throttle for smooth animations
         * @param {Function} func - Function to throttle
         * @returns {Function} RAF throttled function
         */
        rafThrottle(func) {
            let rafId = null;
            return function(...args) {
                if (rafId === null) {
                    rafId = requestAnimationFrame(() => {
                        func.apply(this, args);
                        rafId = null;
                    });
                }
            };
        },

        /**
         * Batch DOM updates to reduce reflows
         * @param {Function} callback - Function containing DOM updates
         */
        batchUpdate(callback) {
            requestAnimationFrame(() => {
                callback();
            });
        }
    };

    // ========================================
    // 3. Virtual Scrolling for Long Lists
    // ========================================

    const VirtualScroll = {
        /**
         * Create a virtual scroll container
         * @param {HTMLElement} container - Container element
         * @param {Array} data - Data array
         * @param {Object} options - Configuration options
         * @returns {Object} Virtual scroll instance
         */
        create(container, data, options = {}) {
            const defaults = {
                itemHeight: 60,          // Height of each item in px
                bufferSize: 5,           // Number of items to render outside viewport
                renderItem: null,        // Function to render item
                onScroll: null           // Scroll callback
            };

            const config = { ...defaults, ...options };
            if (!config.renderItem) {
                console.error('renderItem function is required');
                return null;
            }

            const state = {
                data,
                scrollTop: 0,
                containerHeight: container.clientHeight,
                visibleStart: 0,
                visibleEnd: 0
            };

            // Create inner container
            const inner = document.createElement('div');
            inner.style.height = `${data.length * config.itemHeight}px`;
            inner.style.position = 'relative';
            container.appendChild(inner);

            // Create viewport for visible items
            const viewport = document.createElement('div');
            viewport.style.position = 'absolute';
            viewport.style.top = '0';
            viewport.style.left = '0';
            viewport.style.right = '0';
            inner.appendChild(viewport);

            /**
             * Render visible items
             */
            const render = () => {
                const totalHeight = state.data.length * config.itemHeight;
                const visibleCount = Math.ceil(state.containerHeight / config.itemHeight);

                state.visibleStart = Math.max(0, Math.floor(state.scrollTop / config.itemHeight) - config.bufferSize);
                state.visibleEnd = Math.min(
                    state.data.length,
                    Math.ceil((state.scrollTop + state.containerHeight) / config.itemHeight) + config.bufferSize
                );

                const offsetY = state.visibleStart * config.itemHeight;

                // Clear viewport
                viewport.innerHTML = '';
                viewport.style.transform = `translateY(${offsetY}px)`;

                // Render visible items using document fragment for better performance
                const fragment = document.createDocumentFragment();

                for (let i = state.visibleStart; i < state.visibleEnd; i++) {
                    const item = config.renderItem(state.data[i], i);
                    item.style.position = 'absolute';
                    item.style.top = `${i * config.itemHeight}px`;
                    item.style.left = '0';
                    item.style.right = '0';
                    item.style.height = `${config.itemHeight}px`;
                    fragment.appendChild(item);
                }

                viewport.appendChild(fragment);

                if (config.onScroll) {
                    config.onScroll({
                        startIndex: state.visibleStart,
                        endIndex: state.visibleEnd,
                        scrollTop: state.scrollTop
                    });
                }
            };

            /**
             * Handle scroll event
             */
            const handleScroll = PerformanceUtils.throttle(() => {
                state.scrollTop = container.scrollTop;
                render();
            }, 16); // ~60fps

            /**
             * Update data and re-render
             * @param {Array} newData - New data array
             */
            const updateData = (newData) => {
                state.data = newData;
                inner.style.height = `${newData.length * config.itemHeight}px`;
                render();
            };

            /**
             * Get current viewport state
             * @returns {Object} Viewport state
             */
            const getState = () => ({
                ...state,
                visibleCount: state.visibleEnd - state.visibleStart
            });

            // Initialize
            container.style.overflow = 'auto';
            container.style.height = '100%';
            container.addEventListener('scroll', handleScroll, { passive: true });

            // Initial render
            render();

            // Return public methods
            return {
                render,
                updateData,
                getState,
                destroy: () => {
                    container.removeEventListener('scroll', handleScroll);
                    container.innerHTML = '';
                }
            };
        }
    };

    // ========================================
    // 4. Performance Monitor
    // ========================================

    const PerformanceMonitor = {
        metrics: {},
        observers: [],

        /**
         * Start measuring performance
         * @param {string} label - Metric label
         */
        start(label) {
            this.metrics[label] = performance.now();
        },

        /**
         * End measuring and log performance
         * @param {string} label - Metric label
         * @param {boolean} log - Log to console
         * @returns {number} Duration in ms
         */
        end(label, log = true) {
            if (!this.metrics[label]) {
                console.warn(`No start metric found for: ${label}`);
                return 0;
            }

            const duration = performance.now() - this.metrics[label];
            delete this.metrics[label];

            if (log) {
                console.log(`⏱️ ${label}: ${duration.toFixed(2)}ms`);
            }

            return duration;
        },

        /**
         * Measure function execution time
         * @param {Function} func - Function to measure
         * @param {string} label - Metric label
         * @returns {*} Function result
         */
        measure(func, label) {
            this.start(label);
            const result = func();
            this.end(label);
            return result;
        },

        /**
         * Monitor Core Web Vitals
         * @param {Function} callback - Callback with metrics
         */
        monitorCoreWebVitals(callback) {
            // Observe Largest Contentful Paint (LCP)
            if ('PerformanceObserver' in global) {
                try {
                    const lcpObserver = new PerformanceObserver((list) => {
                        const entries = list.getEntries();
                        const lastEntry = entries[entries.length - 1];
                        callback({ lcp: lastEntry.renderTime || lastEntry.loadTime });
                    });
                    lcpObserver.observe({ entryTypes: ['largest-contentful-paint'] });
                    this.observers.push(lcpObserver);

                    // Observe First Input Delay (FID)
                    const fidObserver = new PerformanceObserver((list) => {
                        const entries = list.getEntries();
                        callback({ fid: entries[0].processingStart - entries[0].startTime });
                    });
                    fidObserver.observe({ entryTypes: ['first-input'] });
                    this.observers.push(fidObserver);

                    // Observe Cumulative Layout Shift (CLS)
                    let clsValue = 0;
                    const clsObserver = new PerformanceObserver((list) => {
                        for (const entry of list.getEntries()) {
                            if (!entry.hadRecentInput) {
                                clsValue += entry.value;
                                callback({ cls: clsValue });
                            }
                        }
                    });
                    clsObserver.observe({ entryTypes: ['layout-shift'] });
                    this.observers.push(clsObserver);
                } catch (e) {
                    console.warn('PerformanceObserver error:', e);
                }
            }
        },

        /**
         * Get memory usage (if available)
         * @returns {Object|null} Memory info
         */
        getMemoryUsage() {
            if (global.performance && global.performance.memory) {
                return {
                    used: Math.round(global.performance.memory.usedJSHeapSize / 1048576),
                    total: Math.round(global.performance.memory.totalJSHeapSize / 1048576),
                    limit: Math.round(global.performance.memory.jsHeapSizeLimit / 1048576)
                };
            }
            return null;
        },

        /**
         * Disconnect all observers
         */
        destroy() {
            this.observers.forEach(observer => observer.disconnect());
            this.observers = [];
        }
    };

    // ========================================
    // 5. Memory Management
    // ========================================

    const MemoryManager = {
        eventListeners: new Map(),
        intervals: new Set(),
        timeouts: new Set(),

        /**
         * Track event listener for cleanup
         * @param {HTMLElement} element - Target element
         * @param {string} event - Event name
         * @param {Function} handler - Event handler
         * @param {Object} options - Event options
         */
        addEventListener(element, event, handler, options) {
            element.addEventListener(event, handler, options);

            const key = `${element}_${event}`;
            if (!this.eventListeners.has(key)) {
                this.eventListeners.set(key, []);
            }
            this.eventListeners.get(key).push({ element, event, handler, options });
        },

        /**
         * Track interval for cleanup
         * @param {Function} callback - Interval callback
         * @param {number} delay - Delay in ms
         * @returns {number} Interval ID
         */
        setInterval(callback, delay) {
            const id = setInterval(callback, delay);
            this.intervals.add(id);
            return id;
        },

        /**
         * Track timeout for cleanup
         * @param {Function} callback - Timeout callback
         * @param {number} delay - Delay in ms
         * @returns {number} Timeout ID
         */
        setTimeout(callback, delay) {
            const id = setTimeout(callback, delay);
            this.timeouts.add(id);
            return id;
        },

        /**
         * Clean up all tracked resources
         */
        cleanup() {
            // Remove event listeners
            this.eventListeners.forEach((listeners) => {
                listeners.forEach(({ element, event, handler, options }) => {
                    element.removeEventListener(event, handler, options);
                });
            });
            this.eventListeners.clear();

            // Clear intervals
            this.intervals.forEach(id => clearInterval(id));
            this.intervals.clear();

            // Clear timeouts
            this.timeouts.forEach(id => clearTimeout(id));
            this.timeouts.clear();
        }
    };

    // ========================================
    // 6. DOM Utilities
    // ========================================

    const DOMUtils = {
        /**
         * Query selector with caching
         * @param {string} selector - CSS selector
         * @param {boolean} useCache - Use cached result
         * @returns {HTMLElement|null} Element
         */
        querySelector(selector, useCache = true) {
            if (!useCache) {
                return document.querySelector(selector);
            }

            if (!this._cache) this._cache = new Map();
            if (!this._cache.has(selector)) {
                this._cache.set(selector, document.querySelector(selector));
            }
            return this._cache.get(selector);
        },

        /**
         * Clear selector cache
         */
        clearCache() {
            if (this._cache) this._cache.clear();
        },

        /**
         * Create element with attributes
         * @param {string} tag - HTML tag
         * @param {Object} attrs - Attributes
         * @param {string} content - Inner content
         * @returns {HTMLElement} Created element
         */
        createElement(tag, attrs = {}, content = '') {
            const el = document.createElement(tag);
            Object.entries(attrs).forEach(([key, value]) => {
                if (key === 'className') {
                    el.className = value;
                } else if (key.startsWith('data-')) {
                    el.dataset[key.slice(5)] = value;
                } else {
                    el.setAttribute(key, value);
                }
            });
            if (content) el.innerHTML = content;
            return el;
        },

        /**
         * Efficiently insert HTML using DocumentFragment
         * @param {string} html - HTML string
         * @param {HTMLElement} parent - Parent element
         */
        insertHTML(html, parent) {
            const template = document.createElement('template');
            template.innerHTML = html.trim();
            parent.appendChild(template.content);
        }
    };

    // ========================================
    // Export to Global
    // ========================================

    global.PerformanceToolkit = {
        LazyLoader,
        PerformanceUtils,
        VirtualScroll,
        PerformanceMonitor,
        MemoryManager,
        DOMUtils
    };

    // Auto-initialize lazy loading
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', () => {
            LazyLoader.init();
        });
    } else {
        LazyLoader.init();
    }

    // Log performance info in development
    if (global.location && global.location.hostname === 'localhost') {
        console.log('🚀 Performance Toolkit loaded');
        PerformanceMonitor.monitorCoreWebVitals((metrics) => {
            console.log('📊 Core Web Vitals:', metrics);
        });
    }

})(window);
