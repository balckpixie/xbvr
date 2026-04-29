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
    const video = this.$refs.vrVideo;

    this.vr.renderer = new THREE.WebGLRenderer({ canvas, antialias: true });
    this.vr.renderer.setPixelRatio(window.devicePixelRatio);
    this.vr.renderer.outputColorSpace = THREE.SRGBColorSpace; 

    this.vr.scene = new THREE.Scene();
    
    // カメラ設定：参考コードと同じ位置
    this.vr.camera = new THREE.PerspectiveCamera(60, 1, 1, 20000);
    this.vr.camera.position.set(100, 0, 0); 
    this.vr.camera.lookAt(0, 0, 0);

    // OrbitControls：拡大縮小（Distance）の制限を適用
    this.vr.controls = new OrbitControls(this.vr.camera, canvas);
    this.vr.controls.rotateSpeed = -0.5;
    this.vr.controls.enableZoom = true;
    this.vr.controls.minDistance = 100; // 最大ズーム
    this.vr.controls.maxDistance = 600; // 最小ズーム
    this.vr.controls.enableDamping = false;
    this.vr.controls.enablePan = false;

    // 全球ジオメトリ
    const geometry = new THREE.SphereGeometry(1000, 60, 40);
    geometry.scale(-1, 1, 1);

    const texture = new THREE.VideoTexture(video);
    texture.colorSpace = THREE.SRGBColorSpace;
    texture.minFilter = texture.magFilter = THREE.LinearFilter;
    
    const material = new THREE.MeshBasicMaterial({ map: texture });
    const mesh = new THREE.Mesh(geometry, material);

    // 【重要】映像の中心（正面）を向かせるための回転
    // カメラが X:100 にあるため、球体を Y 軸に 90 度回転させることで
    // 映像の中心（テクスチャの 0.5 付近）がカメラの正面に来るようになります。
    mesh.rotation.y = Math.PI / 2; 

    this.vr.scene.add(mesh);
    
    this.handleResize();
    this.isReady = true;
    this.animate();
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
    // handleWheel(e) {
    //   e.preventDefault();
      
    //   // 変化の感度を調整（数値が小さいほど滑らかになります）
    //   // 以前の 0.07 から 0.03 に変更し、さらに現在の FOV に応じて感度を変えることで
    //   // ズームインしている時ほどゆっくり変化するようにします。
    //   const sensitivity = 0.01;
    //   const zoomAmount = e.deltaY * sensitivity * (this.vr.camera.fov / 60);
      
    //   // 新しい FOV を計算
    //   const newFov = this.vr.camera.fov + zoomAmount;
      
    //   // 30度から90度の範囲でクランプ（制限）
    //   this.vr.camera.fov = Math.max(30, Math.min(90, newFov));
      
    //   this.vr.camera.updateProjectionMatrix();
    // },
    handleWheel(e) {
    // e.preventDefault();
    // // 前回の回答よりさらに細かい制御（感度 0.02）
    // const sensitivity = 0.02;
    // const zoomAmount = e.deltaY * sensitivity;
    
    // // FOV（画角）によるズーム
    // const newFov = this.vr.camera.fov + zoomAmount;
    // this.vr.camera.fov = Math.max(30, Math.min(90, newFov));
    // this.vr.camera.updateProjectionMatrix();

    // または、参考コードのようにカメラの距離(Distance)を変える場合は以下
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
  /* 100% を基本にしつつ、ウィンドウからはみ出さないように 
     vh（ビューポートの高さ）で上限を強制します。
  */
  height: 100%;
  min-height: 300px;   /* 潰れ防止の最小値 */
  max-height: 75vh;    /* ウィンドウの75%以上には絶対にならないように制限 */
  
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