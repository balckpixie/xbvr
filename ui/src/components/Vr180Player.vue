<template>
  <div class="vr-player-wrapper">
    <div class="vr-canvas-container">
      <canvas ref="vrCanvas" class="vr-canvas"></canvas>
      <div v-if="!isReady" class="vr-loading">
        <div class="loader"></div>
      </div>
    </div>
    <video ref="vrVideo" crossorigin="anonymous" playsinline muted style="display:none"></video>
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
      vr: { renderer: null, scene: null, camera: null, controls: null, animationId: null },
      resizeObserver: null
    };
  },
  watch: {
    fileId: 'updateVideoSource'
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
</style>