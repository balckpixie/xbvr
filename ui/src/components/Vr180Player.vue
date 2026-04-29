<template>
  <div class="vr-player-wrapper">
    <div class="vr-canvas-container" @mousemove="showControls">
      <canvas ref="vrCanvas" class="vr-canvas"></canvas>
      
      <div v-if="!isReady" class="vr-loading">
        <div class="loader"></div>
      </div>

      <div class="vr-controls" :class="{ 'vr-controls-hide': !uiVisible }">
        <div class="control-row seek-bar-row">
          <input type="range" class="seek-bar" step="0.1"
            :min="0" :max="videoDuration" :value="videoCurrentTime"
            @input="onSeek" @mousedown="isSeeking = true" @mouseup="isSeeking = false">
        </div>

        <div class="control-row button-row">
          <div class="left-controls">
            <button @click="togglePlay" class="ctrl-btn">
              <span v-if="paused">▶</span><span v-else>||</span>
            </button>
            <button @click="stopVideo" class="ctrl-btn">■</button>
            <button @click="rewind" class="ctrl-btn">-10s</button>
            <button @click="fastForward" class="ctrl-btn">+10s</button>
            <span class="time-display">{{ formatTime(videoCurrentTime) }} / {{ formatTime(videoDuration) }}</span>
          </div>

          <div class="right-controls">
            <button @click="recenter" class="ctrl-btn">Recenter</button>
            <button @click="toggleMute" class="ctrl-btn">
              <span v-if="isMuted">Mute ON</span><span v-else>Mute OFF</span>
            </button>
            <input type="range" class="volume-bar" min="0" max="1" step="0.1" v-model="volume">
            <button @click="toggleFullScreen" class="ctrl-btn">Fullscreen</button>
          </div>
        </div>
      </div>
    </div>
    <video ref="vrVideo" crossorigin="anonymous" playsinline style="display:none"
      @loadedmetadata="onMetadataLoaded" @timeupdate="onTimeUpdate" @ended="onEnded"></video>
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
      // UI管理
      uiVisible: true,
      uiTimer: null,
      isSeeking: false,
      // ビデオ状態
      paused: true,
      videoCurrentTime: 0,
      videoDuration: 0,
      volume: 1,
      isMuted: true,
      
      isReady: false,
      vr: { renderer: null, scene: null, camera: null, controls: null, animationId: null },
      resizeObserver: null
    };
  },
  watch: {
    fileId: 'updateVideoSource',
    volume(val) {
     if (this.$refs.vrVideo) this.$refs.vrVideo.volume = val;
    }
  },

  mounted() {
    this.initThree();
    this.updateVideoSource();
    
    // 要素自体のサイズ変更を監視（スプリッター対応）
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

      // 1. レンダラー、シーン、カメラの基本設定
      this.initBaseScene(canvas);

      // 2. カメラ操作（ズーム・回転）の設定
      this.initControls(canvas);

      // 3. 180SBS動画用マッピングメッシュの作成
      this.initVRMesh();

      this.handleResize();
      this.isReady = true;
      this.animate();
    },

    /**
     * シーン・レンダラー・カメラの初期化
     */
    initBaseScene(canvas) {
      this.vr.renderer = new THREE.WebGLRenderer({ canvas, antialias: true });
      this.vr.renderer.setPixelRatio(window.devicePixelRatio);
      this.vr.renderer.outputColorSpace = THREE.SRGBColorSpace;

      this.vr.scene = new THREE.Scene();

      this.vr.camera = new THREE.PerspectiveCamera(60, 1, 1, 20000);
      // 初期位置を x: 100 に設定し、中心を向かせる
      this.vr.camera.position.set(100, 0, 0);
      this.vr.camera.lookAt(0, 0, 0);
    },

    /**
     * OrbitControls（拡大縮小・回転）の初期化
     */
    initControls(canvas) {
      this.vr.controls = new OrbitControls(this.vr.camera, canvas);
      this.vr.controls.rotateSpeed = -0.5;
      this.vr.controls.enableZoom = true;
      this.vr.controls.minDistance = 100; // 拡大上限
      this.vr.controls.maxDistance = 800; // 縮小上限
      this.vr.controls.enableDamping = false;
      this.vr.controls.enablePan = false;
    },

    /**
     * 180SBS動画を全球にマッピング（前方：正位置、後方：反転）
     */
    initVRMesh() {
      const video = this.$refs.vrVideo;

      // 1. 360度全球ジオメトリの作成
      const geometry = new THREE.SphereGeometry(1000, 60, 40);
      geometry.scale(-1, 1, 1); // 内側を表示

      // 2. UV座標の計算：180SBS動画の左半分(0.0-0.5)だけを使用するように調整
      const uvs = geometry.attributes.uv;
      for (let i = 0; i < uvs.count; i++) {
        let u = uvs.getX(i); // 球体表面の割合 (0.0〜1.0)

        if (u <= 0.5) {
          // 【前方エリア】球体の 0〜180度
          // 動画の左半分(0.0〜0.5)をそのまま割り当てる
          uvs.setX(i, u);
        } else {
          // 【後方エリア】球体の 180〜360度
          // 動画の左半分を反転(0.5〜0.0)させて割り当てる
          uvs.setX(i, 1.0 - u);
        }
      }
      uvs.needsUpdate = true;

      // 3. テクスチャとマテリアルの作成
      const texture = new THREE.VideoTexture(video);
      texture.colorSpace = THREE.SRGBColorSpace;
      texture.wrapS = THREE.ClampToEdgeWrapping;
      texture.minFilter = texture.magFilter = THREE.LinearFilter;

      const material = new THREE.MeshBasicMaterial({ map: texture });
      const mesh = new THREE.Mesh(geometry, material);

      // 4. 正面合わせの回転
      // SBSの左半分の中央（0.25地点）が正面に来るよう調整
      mesh.rotation.y = Math.PI / 2;

      this.vr.scene.add(mesh);
    },

    updateVideoSource() {
      if (this.$refs.vrVideo && this.fileId) {
        this.$refs.vrVideo.src = `/api/dms/file/${this.fileId}?dnt=true`;
        this.$refs.vrVideo.play().catch(() => {});
      }
    },
    animate() {
      this.vr.animationId = requestAnimationFrame(this.animate);
      if (this.vr.controls) this.vr.controls.update();
      if (this.vr.renderer) this.vr.renderer.render(this.vr.scene, this.vr.camera);
    },
    handleResize() {
      const container = this.$el;
      if (!container || !this.vr.renderer) return;

      // 親要素の現在のサイズを取得
      let width = container.clientWidth;
      let height = container.clientHeight;

      // 安全策：ウィンドウの高さの 90% を超える場合は制限する
      // （CSSの max-height が効いていれば、基本的にはここを通ることはありません）
      const maxHeight = window.innerHeight * 0.8;
      if (height > maxHeight) {
        height = maxHeight;
      }

      if (width === 0 || height === 0) return;

      // レンダラーのサイズを更新（第3引数を false にして CSS との競合を防ぐ）
      this.vr.renderer.setSize(width, height, false);
      this.vr.camera.aspect = width / height;
      this.vr.camera.updateProjectionMatrix();
    },

    handleWheel(e) {
    // this.vr.controls.handleMouseWheel(e); // OrbitControlsのデフォルトズームを使う場合
    },

    dispose() {
      cancelAnimationFrame(this.vr.animationId);
      if (this.vr.renderer) {
        this.vr.renderer.dispose();
        this.vr.renderer.forceContextLoss();
      }
    },

    // ① 時間フォーマット (sec2timestr の移植)
    formatTime(sec) {
      if (!sec || isNaN(sec)) return "00:00:00";
      const h = Math.floor(sec / 3600).toString().padStart(2, '0');
      const m = Math.floor((sec % 3600) / 60).toString().padStart(2, '0');
      const s = Math.floor(sec % 60).toString().padStart(2, '0');
      return `${h}:${m}:${s}`;
    },
    // メタデータ読み込み
    onMetadataLoaded() {
      this.videoDuration = this.$refs.vrVideo.duration;
    },
    // ⑤ シーク更新
    onTimeUpdate() {
      if (!this.isSeeking) {
        this.videoCurrentTime = this.$refs.vrVideo.currentTime;
      }
      this.paused = this.$refs.vrVideo.paused;
    },
    // ② 一時停止／再生
    togglePlay() {
      const video = this.$refs.vrVideo;
      if (video.paused) video.play(); else video.pause();
    },
    // ③ 停止
    stopVideo() {
      const video = this.$refs.vrVideo;
      video.pause();
      video.currentTime = 0;
    },
    // ④ 巻き戻し / ⑥ 早送り
    rewind() { this.$refs.vrVideo.currentTime -= 10; },
    fastForward() { this.$refs.vrVideo.currentTime += 10; },
    // ⑤ シーク操作
    onSeek(e) {
      const val = parseFloat(e.target.value);
      this.videoCurrentTime = val;
      this.$refs.vrVideo.currentTime = val;
    },
    // ⑧ ミュート
    toggleMute() {
      this.isMuted = !this.isMuted;
      this.$refs.vrVideo.muted = this.isMuted;
    },
    // ⑨ リセンター (参考ソースの set(90, 0, 0.01) を適用)
    recenter() {
      if (this.vr.camera) {
        this.vr.camera.position.set(100, 0, 0); 
        this.vr.controls.reset();
      }
    },
    // ⑩ 全画面表示
    toggleFullScreen() {
      const el = this.$el;
      if (!document.fullscreenElement) {
        el.requestFullscreen().catch(err => console.error(err));
      } else {
        document.exitFullscreen();
      }
    },
    // UIの自動非表示
    showControls() {
      this.uiVisible = true;
      clearTimeout(this.uiTimer);
      this.uiTimer = setTimeout(() => {
        this.uiVisible = false;
      }, 3000);
    }
  }
};
</script>

<style scoped>
.vr-player-wrapper {
  position: relative; /* 子の absolute の基準点 */
  width: 100%;
  height: 100%;
  min-height: 300px;   /* 潰れ防止の最小値 */
  /* max-height: 83vh; */    /* ウィンドウの75%以上には絶対にならないように制限 */
  
  margin: 0;
  padding: 0;
  background-color: #000;
  overflow: hidden;    /* はみ出しを物理的にカット */
}

.vr-canvas-container {
  /* 絶対配置にすることで、この要素が親（.vr-player-wrapper）の
     サイズを押し広げる「フィードバックループ」を完全に遮断します。
  */
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  width: 100%;
  height: 100%;
  background: #000;
  cursor: grab;
}

.vr-canvas {
  /* Three.js が管理するサイズをそのまま表示 */
  display: block;
  width: 100% !important;
  height: 100% !important;
}

.vr-canvas-container:active {
  cursor: grabbing;
}

.vr-loading {
  position: absolute;
  top: 0; left: 0; width: 100%; height: 100%;
  display: flex; align-items: center; justify-content: center;
  background: rgba(0,0,0,0.8); z-index: 10;
}

.vr-controls {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  background: linear-gradient(transparent, rgba(0,0,0,0.7));
  color: white;
  padding: 10px;
  transition: opacity 0.5s;
  z-index: 100;
}
.vr-controls-hide {
  opacity: 0;
  pointer-events: none;
}
.control-row {
  display: flex;
  align-items: center;
  gap: 15px;
}
.button-row {
  justify-content: space-between;
  margin-top: 5px;
}
.left-controls, .right-controls {
  display: flex;
  align-items: center;
  gap: 10px;
}
.seek-bar {
  width: 100%;
  cursor: pointer;
}
.ctrl-btn {
  background: rgba(255,255,255,0.2);
  border: none;
  color: white;
  padding: 5px 10px;
  border-radius: 4px;
  cursor: pointer;
}
.ctrl-btn:hover { background: rgba(255,255,255,0.4); }
.time-display { font-size: 14px; font-family: monospace; }
.volume-bar { width: 60px; }

</style>