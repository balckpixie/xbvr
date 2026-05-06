<template>
  <div class="vr-player-wrapper">
    <div class="vr-canvas-container"
      @mousedown="onCanvasMouseDown"
      @mousemove="onCanvasMouseMove"
      @mouseup="onCanvasMouseUp"
      >
      <canvas ref="vrCanvas" class="vr-canvas"></canvas>
      
      <div v-if="!isReady" class="vr-loading">
        <div class="loader"></div>
      </div>

      <div
        class="vr-controls"
        :class="{ 'vr-controls-hide': !uiVisible }"
        @mousedown.stop
        @mouseup.stop
      >
        
        <div class="control-row seek-bar-row" @mousedown.stop>
          <!-- サムネイルプレビュー: hoverTime を表示 -->
          <div 
            v-if="spriteConfig" 
            class="thumbnail-preview" 
            :style="thumbnailStyle"
          >
            <!-- 画像エリアのすぐ下に時間を配置 -->
            <div class="thumbnail-time-container">
              <span class="thumbnail-time">{{ formatTime(hoverTime) }}</span>
            </div>
          </div>

          <input 
            type="range" 
            class="seek-bar" 
            step="0.001"
            :min="0" 
            :max="duration" 
            :value="currentTime"
            @mousedown="onSeekStart"
            @input="onSeekInput"
            @mouseup="onSeekEnd"
            @mousemove="updateThumbnail"
            @mouseleave="thumbnailStyle.display = 'none'"
          >
        </div>

        <div class="control-row button-row" @mousedown.stop>
          <div class="left-controls">
            <button class="ctrl-btn" @click="togglePlay">
              <span class="material-icons">{{ isPaused ? 'play_arrow' : 'pause' }}</span>
            </button>
            <button class="ctrl-btn" @click="onStop">
              <span class="material-icons">stop</span>
            </button>
            <button class="ctrl-btn" @click="onRewind">
              <span class="material-icons">replay_10</span>
            </button>
            <button class="ctrl-btn" @click="onFastForward">
              <span class="material-icons">forward_10</span>
            </button>
            <span class="time-display">
              {{ formatTime(currentTime) }} / {{ formatTime(duration) }}
            </span>
          </div>

          <div class="right-controls">
            <button class="ctrl-btn" @click="onRecenter" title="Recenter">
              <span class="material-icons">filter_center_focus</span>
            </button>
            <button class="ctrl-btn" :class="{ 'is-active': isMuted }" @click="onToggleMute">
              <span class="material-icons">{{ isMuted ? 'volume_off' : 'volume_up' }}</span>
            </button>
            <input type="range" class="volume-bar" min="0" max="1" step="0.1" @input="onVolumeChange">
            <button class="ctrl-btn" @click="onToggleFullScreen">
              <span class="material-icons">fullscreen</span>
            </button>
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

    <!-- ローディングオーバーレイ -->
    <div v-if="isBuffering" class="loading-overlay">
      <div class="spinner"></div>
    </div>

  </div>
</template>

<script>
import * as THREE from 'three';
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls';

export default {
  props: {
    // fileId: { type: Number, default: null } // 廃止
    file: { type: Object, default: null }      // 新設: Details.vue 等から渡されるオブジェクト
  },
  data() {
    return {
      isReady: false,
      isPaused: true,
      isMuted: true,
      isSeeking: false, // シークバーをドラッグ中かどうか
      isBuffering: false,  // 動画ロード（バッファリング）中
      isDragging: false,
      mouseMoved: false,
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
      resizeObserver: null,
      //for Sprite
      hoverTime: 0,
      spriteConfig: null, // spriteParamsの結果を保持
      thumbnailStyle: {
        display: 'none',
        left: '0px',
        width: '0px',
        height: '0px',
        backgroundImage: '',
        backgroundPosition: '0px 0px',
        backgroundSize: '0px 0px'
      },
    };
  },
  watch: {
    //fileId: 'updateVideoSource'
    file: {
      immediate: true,
      handler(newFile, oldFile) {
        if (newFile && (!oldFile || newFile.id !== oldFile.id)) {
          this.updateVideoSource(newFile);
        }
      }
    }
  },
  mounted() {
    this.initThree();
    if (this.file) {
      this.updateVideoSource(this.file);
    }
    
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
    dispose() {
      cancelAnimationFrame(this.vr.animationId);
      if (this.vr.renderer) {
        this.vr.renderer.dispose();
        this.vr.renderer.forceContextLoss();
      }
    },

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
      // --- なめらか動作の設定 ---
      // 慣性（ズームや回転を止めた後の余韻）を有効にする
      this.vr.controls.enableDamping = true;
      // 慣性の強さ (0.0 ～ 1.0)。値が小さいほど「ぬるっ」と動きます
      this.vr.controls.dampingFactor = 0.05;
      // ズームの速さも調整可能（必要に応じて）
      this.vr.controls.zoomSpeed = 2.0;
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
      mesh.rotation.y = (-1) * Math.PI / 2;
      this.vr.scene.add(mesh);
    },

    updateVideoSource(file) {
      const video = this.$refs.vrVideo;
      if (!video || !file) return;

      // 1. スプライトパラメータを計算して保持
      this.updateSpriteParams(file);

      // 2. 動画ソースの設定と再生
      video.src = `/api/dms/file/${file.id}?dnt=true`;
      video.load(); // ソース変更時は念のため明示的にロード
      video.play().catch(() => {
        console.log("Autoplay blocked or video not ready.");
      });
    },

    togglePlay() {
      const video = this.$refs.vrVideo;
      if (video.paused) {
        video.play();
      } else {
        video.pause();
      }
    },
    onCanvasMouseDown() {
      this.isDragging = true;
      this.mouseMoved = false; // 押し下げた瞬間は移動していない
    },
    onCanvasMouseMove() {
      if (this.isDragging) {
        this.mouseMoved = true; // マウスが動いたのでドラッグと判定
      }
      this.handleMouseMove(); // UI表示用
    },
    onCanvasMouseUp() {
      this.isDragging = false;
      // マウスが動いていない（純粋なクリック）場合のみ再生/一時停止
      if (!this.mouseMoved) {
        this.togglePlay();
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

    // シークバーを触り始めた時
    onSeekStart() {
      this.isSeeking = true;
      this.handleMouseMove(); // UIが消えないように
    },

    // シークバーを動かしている最中 (プレビュー感覚で時間を更新)
    onSeekInput(e) {
      this.currentTime = parseFloat(e.target.value);
      this.handleMouseMove();
    },

 onSeekEnd(e) {
  const video = this.$refs.vrVideo;
  const seekBar = this.$el.querySelector('.seek-bar');
  if (!video || !seekBar || !Number.isFinite(this.duration)) {
    this.isSeeking = false;
    return;
  }

  // --- updateThumbnail と同じ計算ロジックを適用 ---
  const rect = seekBar.getBoundingClientRect();
  let x = e.clientX - rect.left;
  x = Math.max(0, Math.min(x, rect.width));

  // ピクセルベースで時間を算出
  const seekValue = (x * this.duration) / rect.width;

  if (Number.isFinite(seekValue)) {
    video.currentTime = seekValue;
    this.currentTime = seekValue;
    console.log(`Seeked (Sync) to: ${this.formatTime(seekValue)}`);
  }

  this.isSeeking = false;
  this.handleMouseMove();
},

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

        const setBufferingTrue = () => { this.isBuffering = true; };
        const setBufferingFalse = () => { this.isBuffering = false; };

        video.addEventListener('waiting', setBufferingTrue);
        video.addEventListener('seeking', setBufferingTrue);
        video.addEventListener('playing', setBufferingFalse);
        video.addEventListener('seeked', setBufferingFalse);

        this.$once('hook:beforeDestroy', () => {
          video.removeEventListener('waiting', setBufferingTrue);
          video.removeEventListener('seeking', setBufferingTrue);
          video.removeEventListener('playing', setBufferingFalse);
          video.removeEventListener('seeked', setBufferingFalse);
        });
      }
    },

    onTimeUpdate() {
      const video = this.$refs.vrVideo;
      if (video && !this.isSeeking) {
        // シーク操作中でなければ、現在の再生位置を更新
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
      
      // 全画面表示中かどうかを判定
      const isFull = !!document.fullscreenElement;

      let width, height;

      if (isFull) {
        // 全画面時はブラウザの表示領域いっぱいに設定
        width = window.innerWidth;
        height = window.innerHeight;
      } else {
        width = container.clientWidth;
        height = container.clientHeight;
        const maxHeight = window.innerHeight * 0.8;
        if (height > maxHeight) height = maxHeight;
      }
      if (width === 0 || height === 0) return;

      this.vr.renderer.setSize(width, height, false);
      this.vr.camera.aspect = width / height;
      this.vr.camera.updateProjectionMatrix();
    },

    updateSpriteParams(file) {
      if (!file || !file.thumbnail_parameters) {
        this.spriteConfig = null;
        return;
      }

      const thumbnailUrl = '/api_custom/thumbnail/image/' + file.id;
      const parsed = (typeof file.thumbnail_parameters === 'string')
        ? JSON.parse(file.thumbnail_parameters)
        : file.thumbnail_parameters;

      let tileHeight = parsed.resolution;
      if (file.projection === 'flat') {
        tileHeight = (file.video_height / file.video_width) * parsed.resolution;
      }

      // --- 画像サイズを取得する処理 ---
      const img = new Image();
      img.src = thumbnailUrl;
      
      img.onload = () => {
        // 画像全体のサイズを取得
        const fullWidth = img.naturalWidth;
        const fullHeight = img.naturalHeight;

        // 画像サイズから列数・行数を逆算
        const columns = Math.floor(fullWidth / parsed.resolution);
        const rows = Math.floor(fullHeight / tileHeight);

        // 確定した情報を保持
        this.spriteConfig = {
          url: thumbnailUrl,
          duration: file.duration,
          start: parsed.start,
          interval: parsed.interval,
          width: parsed.resolution, // 1コマの幅
          height: tileHeight,       // 1コマの高さ
          columns: columns,         // 計算された列数
          totalWidth: fullWidth,    // 画像全体の幅
          totalHeight: fullHeight   // 画像全体の高さ
        };
      };

      img.onerror = () => {
        console.error("サムネイル画像の読み込みに失敗しました。");
        this.spriteConfig = null;
      };
    },


    updateThumbnail(e) {
      const seekBar = this.$el.querySelector('.seek-bar');
      const config = this.spriteConfig;
      if (!seekBar || !config || !this.duration) return;

      const rect = seekBar.getBoundingClientRect();
      let seekBarPosX = e.clientX - rect.left;
      seekBarPosX = Math.max(0, Math.min(seekBarPosX, rect.width));
      
      // --- はみ出し防止ロジックの追加 ---
      const controlsPadding = 20;
      const halfThumbWidth = config.width / 2;
      let displayLeft = seekBarPosX;

      // 左端の制限: 半分より左に行こうとしたら、左端（halfWidth）で止める
      if (seekBarPosX < halfThumbWidth) {
        displayLeft = halfThumbWidth + controlsPadding;
      } 
      // 右端の制限: 右端から半分より右に行こうとしたら、右端（width - halfWidth）で止める
      else if (seekBarPosX > rect.width + controlsPadding - halfThumbWidth) {
        displayLeft = rect.width + controlsPadding - halfThumbWidth;
      }

      this.hoverTime = parseFloat(seekBarPosX * this.duration / rect.width);
      console.log(`Mouse X: ${seekBarPosX}px, Hover Time: ${this.formatTime(this.hoverTime)}`);
      
      // マウス位置から求めた時間に、スプライトの開始オフセットを考慮
      // 微小な誤差（0.0001）を加えることで、境界での切り捨てミスを防ぐ
      const adjustedTime = Math.max(0, this.hoverTime - config.start);
      const spriteIndex = Math.floor((adjustedTime + 0.0001) / config.interval);

      // 総コマ数を超えないようにガード
      const maxIndex = (config.columns * (config.rows || config.columns)) - 1;
      const safeIndex = Math.max(0, Math.min(spriteIndex, maxIndex));

      const col = safeIndex % config.columns;
      const row = Math.floor(safeIndex / config.columns);

      this.thumbnailStyle = {
        display: 'block',
        left: `${displayLeft}px`,
        width: `${config.width}px`,
        height: `${config.height}px`,
        backgroundImage: `url(${config.url})`,
        // 1コマのサイズ分だけ背景をマイナス方向にずらす
        backgroundPosition: `-${col * config.width}px -${row * config.height}px`,
        // 重要：背景全体のサイズを「1コマの幅 × 列数」に指定して拡大させる
        backgroundSize: `${config.totalWidth}px ${config.totalHeight}px`,
        opacity: 1,
        pointerEvents: 'none'
      };
    },
  }
};
</script>

<style scoped>
@import url('https://fonts.googleapis.com/icon?family=Material+Icons');

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
.seek-bar::-webkit-slider-thumb {
  -webkit-appearance: none;
  width: 16px;
  height: 16px;
}
.seek-bar {
  width: 100%;
  margin: 0;        /* 隙間の原因になるため 0 に */
  padding: 0;       /* 隙間の原因になるため 0 に */
  cursor: pointer;
}

/* ボタンの行 */
.button-row { justify-content: space-between; }
.left-controls, .right-controls { display: flex; align-items: center; gap: 10px; }

.ctrl-btn {
  background: transparent; /* 背景をスッキリさせる */
  color: white;
  border: none; /* 枠線を消してモダンに */
  padding: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: transform 0.1s ease, color 0.2s;
}

.ctrl-btn:hover {
  color: #00e5ff; /* ホバー時にアクセントカラー */
  background: rgba(255, 255, 255, 0.1);
  border-radius: 50%; /* 円形のホバーエフェクト */
}


.time-display { font-size: 13px; font-family: monospace; }

/* ボリュームバーを少しスリムに */
.volume-bar {
  width: 80px;
  height: 4px;
  cursor: pointer;
  accent-color: #00e5ff; /* スライダーの色も合わせる */
}

/* スプライト画像を表示するメインコンテナ */
/* スプライト画像を表示するメインコンテナ */
.thumbnail-preview {
  position: absolute;
  pointer-events: none !important; 
  /* 1. 全体をさらに上に浮かせる (85px から 110px 程度へ) */
  bottom: 110px; 
  left: 0;
  transform: translateX(-50%);
  border: 2px solid rgba(255, 255, 255, 0.8);
  border-radius: 8px;
  z-index: 1000;
  display: none;
  box-sizing: border-box;
  overflow: visible; /* 時間表示を枠外に出すために必須 */
  box-shadow: 0 4px 15px rgba(0, 0, 0, 0.6);
}

/* 時間表示のコンテナ */
.thumbnail-time-container {
  position: absolute;
  /* 2. 画像の下端から少し隙間を空けて配置 (例: -30px) */
  bottom: -32px; 
  left: 50%;
  transform: translateX(-50%);
  width: 100%;
  text-align: center;
}

/* 時間テキスト自体のスタイル */
.thumbnail-time {
  background: rgba(0, 0, 0, 0.7); /* 少し濃くして視認性アップ */
  color: #fff;
  padding: 4px 10px; /* 少し大きくして画像に寄せすぎない */
  border-radius: 4px;
  font-size: 14px;
  font-weight: bold;
  font-family: monospace; /* 数字の幅を一定にする[cite: 1] */
}

.loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  background-color: rgba(0, 0, 0, 0.2); /* 軽く暗くする */
  pointer-events: none; /* 下の要素のクリックを邪魔しない */
  z-index: 2000;
}

/* シンプルなスピナーの例 */
.spinner {
  width: 50px;
  height: 50px;
  border: 5px solid rgba(255, 255, 255, 0.3);
  border-top: 5px solid #fff;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}


.material-icons {
  font-size: 24px; /* アイコンの基本サイズ */
}


</style>