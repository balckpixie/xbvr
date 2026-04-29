<template>
  <div class="vr-player-wrapper">
    <div class="vr-canvas-container" @mousemove="handleMouseMove">
      <canvas ref="vrCanvas" class="vr-canvas"></canvas>
      
      <div v-if="!isReady" class="vr-loading">
        <div class="loader"></div>
      </div>

      <div class="vr-controls" :class="{ 'vr-controls-hide': !uiVisible }">
        
        <div class="control-row seek-bar-row" @mousedown.stop>
          <input type="range" class="seek-bar" min="0" max="100" value="0" @input="onSeek">
        </div>

        <div class="control-row button-row" @mousedown.stop>
          <div class="left-controls">
            <button class="ctrl-btn" @click="togglePlay">
              <span v-if="isPaused">▶</span><span v-else>||</span>
            </button>
            <button class="ctrl-btn" @click="onStop">■</button>
            <button class="ctrl-btn" @click="onRewind">⏪</button>
            <button class="ctrl-btn" @click="onFastForward">⏩</button>
            <span class="time-display">
              {{ formatTime(currentTime) }} / {{ formatTime(duration) }}
            </span>
          </div>

          <div class="right-controls">
            <button class="ctrl-btn" @click="onRecenter">Recenter</button>
            <button class="ctrl-btn" :class="{ 'is-active': isMuted }" @click="onToggleMute">
              <span v-if="isMuted">🔇 Muted</span>
              <span v-else>🔊 Mute</span>
            </button>
            <input type="range" class="volume-bar" min="0" max="1" step="0.1" @input="onVolumeChange">
            <button class="ctrl-btn" @click="onToggleFullScreen">Full</button>
          </div>
        </div>
      </div>
    </div>

    <video 
      ref="vrVideo" 
      crossorigin="anonymous" 
      playsinline 
      muted 
      style="display:none"
      @play="isPaused = false"
      @pause="isPaused = true"
      @loadedmetadata="onLoadedMetadata"
      @timeupdate="onTimeUpdate"
    ></video>
  </div>
</template>

<script>
import * as THREE from 'three';
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls';

export default {
  props: {
    fileId: { type: Number, default: null }
  },
  data() {
    return {
      isReady: false,
      isPaused: true,
      isMuted: true,
      uiVisible: true,  // UIの表示状態
      uiTimer: null,    // 非表示用タイマー
      currentTime: 0,
    duration: 0,
      vr: { 
        renderer: null, 
        scene: null, 
        camera: null, 
        controls: null, 
        animationId: null,
        videoTexture: null // テクスチャ更新用に保持
      },
      resizeObserver: null
    };
  },
  watch: {
    fileId: 'updateVideoSource'
  },
  mounted() {
    this.initThree();
    this.updateVideoSource();
    
    this.resizeObserver = new ResizeObserver(() => {
      this.handleResize();
    });
    this.resizeObserver.observe(this.$el);
    
    window.addEventListener('resize', this.handleResize);
  },
  beforeDestroy() {
    window.removeEventListener('resize', this.handleResize);
    if (this.resizeObserver) this.resizeObserver.disconnect();
    this.dispose();
  },

  methods: {
    initThree() {
      const canvas = this.$refs.vrCanvas;
      this.initBaseScene(canvas);
      this.initControls(canvas);
      this.initVRMesh();
      this.handleResize();
      this.isReady = true;
      this.animate();
    },

    initBaseScene(canvas) {
      this.vr.renderer = new THREE.WebGLRenderer({ canvas, antialias: true });
      this.vr.renderer.setPixelRatio(window.devicePixelRatio);
      this.vr.renderer.outputColorSpace = THREE.SRGBColorSpace;
      this.vr.scene = new THREE.Scene();
      this.vr.camera = new THREE.PerspectiveCamera(60, 1, 1, 20000);
      this.vr.camera.position.set(100, 0, 0);
      this.vr.camera.lookAt(0, 0, 0);
    },

    initControls(canvas) {
      this.vr.controls = new OrbitControls(this.vr.camera, canvas);
      this.vr.controls.rotateSpeed = -0.5;
      this.vr.controls.enableZoom = true;
      this.vr.controls.minDistance = 100;
      this.vr.controls.maxDistance = 800;
      this.vr.controls.enableDamping = false;
      this.vr.controls.enablePan = false;
    },

    initVRMesh() {
      const video = this.$refs.vrVideo;
      const geometry = new THREE.SphereGeometry(1000, 60, 40);
      geometry.scale(-1, 1, 1);

      const uvs = geometry.attributes.uv;
      for (let i = 0; i < uvs.count; i++) {
        let u = uvs.getX(i);
        if (u <= 0.5) {
          uvs.setX(i, u);
        } else {
          uvs.setX(i, 1.0 - u);
        }
      }
      uvs.needsUpdate = true;

      const texture = new THREE.VideoTexture(video);
      texture.colorSpace = THREE.SRGBColorSpace;
      texture.wrapS = THREE.ClampToEdgeWrapping;
      texture.minFilter = texture.magFilter = THREE.LinearFilter;
      
      // テクスチャを保持しておく
      this.vr.videoTexture = texture;

      const material = new THREE.MeshBasicMaterial({ map: texture });
      const mesh = new THREE.Mesh(geometry, material);
      mesh.rotation.y = Math.PI / 2;
      this.vr.scene.add(mesh);
    },

    updateVideoSource() {
      const video = this.$refs.vrVideo;
      if (video && this.fileId) {
        video.src = `/api/dms/file/${this.fileId}?dnt=true`;
        video.play().catch(() => {});
      }
    },

    togglePlay() {
      const video = this.$refs.vrVideo;
      if (video.paused) {
        video.play();
      } else {
        video.pause();
      }
    },

    handleMouseMove() {
      this.uiVisible = true;
      if (this.uiTimer) clearTimeout(this.uiTimer);
      this.uiTimer = setTimeout(() => {
        this.uiVisible = false;
      }, 3000);
    },

    togglePlay() {
      const video = this.$refs.vrVideo;
      if (video.paused) video.play(); else video.pause();
      this.handleMouseMove();
    },

    onStop() {
      const video = this.$refs.vrVideo;
      if (video) {
        video.pause();
        video.currentTime = 0;
      }
      this.handleMouseMove(); // UI表示タイマーをリセット
    },

    onRewind() {
      const video = this.$refs.vrVideo;
      if (video) {
        video.currentTime = Math.max(0, video.currentTime - 10);
      }
      this.handleMouseMove();
    },

    onFastForward() {
      const video = this.$refs.vrVideo;
      if (video) {
        video.currentTime = Math.min(video.duration, video.currentTime + 10);
      }
      this.handleMouseMove();
    },

    onRecenter() {
      if (this.vr.controls && this.vr.camera) {
        // OrbitControls のターゲットと回転をリセット
        this.vr.controls.reset();
        
        // 参考ソースの初期座標 (100, 0, 0) にカメラを再配置
        this.vr.camera.position.set(100, 0, 0);
        this.vr.camera.lookAt(0, 0, 0);
      }
      this.handleMouseMove();
    },

    onToggleMute() {
      const video = this.$refs.vrVideo;
      if (video) {
        video.muted = !video.muted;
        // データの isMuted 状態も更新（UI表示用）
        this.isMuted = video.muted;
      }
      this.handleMouseMove();
    },

    onToggleFullScreen() {
      // canvas コンテナ要素を全画面化の対象とする
      const container = this.$el.querySelector('.vr-canvas-container');
      
      if (!document.fullscreenElement) {
        // 全画面開始
        if (container.requestFullscreen) {
          container.requestFullscreen();
        } else if (container.webkitRequestFullscreen) {
          container.webkitRequestFullscreen(); // Safari用
        }
      } else {
        // 全画面解除
        if (document.exitFullscreen) {
          document.exitFullscreen();
        }
      }
      this.handleMouseMove();
    },

    onSeek(e) { /* ⑤ シークバー操作 */ },
    onVolumeChange(e) { /* ⑦ 音量変更 */ },

    // 秒(数)を 00:00 形式の文字列に変換する
    formatTime(seconds) {
      if (!seconds || isNaN(seconds)) return '00:00';
      const min = Math.floor(seconds / 60);
      const sec = Math.floor(seconds % 60);
      return `${min.toString().padStart(2, '0')}:${sec.toString().padStart(2, '0')}`;
    },

    // ビデオのメタデータが読み込まれたら総時間を取得
    onLoadedMetadata() {
      const video = this.$refs.vrVideo;
      if (video) {
        this.duration = video.duration;
      }
    },

    // 再生位置が更新されるたびに現在の時間を取得
    onTimeUpdate() {
      const video = this.$refs.vrVideo;
      if (video) {
        this.currentTime = video.currentTime;
      }
    },

    animate() {
      this.vr.animationId = requestAnimationFrame(this.animate);
      
      // ビデオフレームの更新を強制
      if (this.vr.videoTexture) {
        this.vr.videoTexture.needsUpdate = true;
      }

      if (this.vr.controls) this.vr.controls.update();
      if (this.vr.renderer) this.vr.renderer.render(this.vr.scene, this.vr.camera);
    },

    handleResize() {
      const container = this.$el;
      if (!container || !this.vr.renderer) return;
      let width = container.clientWidth;
      let height = container.clientHeight;
      const maxHeight = window.innerHeight * 0.8;
      if (height > maxHeight) height = maxHeight;
      if (width === 0 || height === 0) return;

      this.vr.renderer.setSize(width, height, false);
      this.vr.camera.aspect = width / height;
      this.vr.camera.updateProjectionMatrix();
    },

    dispose() {
      cancelAnimationFrame(this.vr.animationId);
      if (this.vr.renderer) {
        this.vr.renderer.dispose();
        this.vr.renderer.forceContextLoss();
      }
    }
  }
};
</script>

<style scoped>
.vr-player-wrapper {
  position: relative;
  width: 100%;
  height: 100%;
  min-height: 300px;
  background-color: #000;
  overflow: hidden;
}

.vr-canvas-container {
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  width: 100%; height: 100%;
  background: #000;
  cursor: grab;
}

.vr-canvas {
  display: block;
  width: 100% !important;
  height: 100% !important;
}

/* UIのコンテナ: 透明部分はクリックを透過させる */
.simple-controls {
  position: absolute;
  bottom: 20px;
  left: 20px;
  z-index: 100;
  pointer-events: none; 
}

/* ボタン自体: クリックを有効にする */
.play-btn {
  pointer-events: auto;
  background: rgba(0, 0, 0, 0.6);
  color: white;
  border: 1px solid #fff;
  padding: 10px 20px;
  cursor: pointer;
  border-radius: 4px;
  font-size: 14px;
}

.play-btn:hover {
  background: rgba(255, 255, 255, 0.2);
}

.vr-loading {
  position: absolute;
  top: 0; left: 0; width: 100%; height: 100%;
  display: flex; align-items: center; justify-content: center;
  background: rgba(0,0,0,0.8); z-index: 10;
}


.vr-controls {
  position: absolute;
  bottom: 0; left: 0; right: 0;
  padding: 40px 20px 20px;
  background: linear-gradient(transparent, rgba(0, 0, 0, 0.8));
  color: white;
  z-index: 100;
  pointer-events: none; /* 透明エリアは透過 */
  transition: opacity 0.5s ease;
}

.vr-controls-hide { opacity: 0; }

.control-row {
  display: flex;
  align-items: center;
  pointer-events: auto; /* UIパーツは操作有効 */
}

/* シークバーの行 */
.seek-bar-row { margin-bottom: 10px; }
.seek-bar { width: 100%; cursor: pointer; }

/* ボタンの行 */
.button-row { justify-content: space-between; }
.left-controls, .right-controls { display: flex; align-items: center; gap: 10px; }

.ctrl-btn {
  background: rgba(255, 255, 255, 0.1);
  color: white;
  border: 1px solid rgba(255, 255, 255, 0.5);
  padding: 5px 12px;
  cursor: pointer;
  border-radius: 4px;
  font-size: 12px;
}

.ctrl-btn:hover { background: rgba(255, 255, 255, 0.3); }

.time-display { font-size: 13px; font-family: monospace; }

.volume-bar { width: 60px; cursor: pointer; }
</style>