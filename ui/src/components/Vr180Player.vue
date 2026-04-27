<template>
  <div class="vr-player-wrapper" ref="vrWrapper">
    <div class="vr-canvas-container" ref="vrContainer">
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
      this.vr.camera = new THREE.PerspectiveCamera(60, 1, 1, 2000);
      this.vr.camera.position.set(0, 0, 0.1);

      this.vr.controls = new OrbitControls(this.vr.camera, canvas);
      this.vr.controls.rotateSpeed = -0.5;
      this.vr.controls.enableZoom = false;
      this.vr.controls.enableDamping = true;

      // VR180 ジオメトリ
      const geometry = new THREE.SphereGeometry(1000, 60, 40, -Math.PI / 2, Math.PI, 0, Math.PI);
      geometry.scale(-1, 1, 1);
      const uvs = geometry.attributes.uv;
      for (let i = 0; i < uvs.count; i++) { uvs.setX(i, uvs.getX(i) * 0.5); }
      uvs.needsUpdate = true;

      const texture = new THREE.VideoTexture(video);
      texture.colorSpace = THREE.SRGBColorSpace;
      
      const material = new THREE.MeshBasicMaterial({ map: texture });
      this.vr.scene.add(new THREE.Mesh(geometry, material));

      canvas.addEventListener('wheel', this.handleWheel, { passive: false });
      
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

      const width = container.clientWidth;
      const height = container.clientHeight;

      if (width === 0 || height === 0) return;

      this.vr.renderer.setSize(width, height, false);
      this.vr.camera.aspect = width / height;
      this.vr.camera.updateProjectionMatrix();
    },
    handleWheel(e) {
      e.preventDefault();
      this.vr.camera.fov = Math.max(30, Math.min(90, this.vr.camera.fov + e.deltaY * 0.07));
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
  /* 表示を安定させるため、以前の 65vh を「基準の高さ」として設定しつつ、
     親要素がそれ以上の高さを持つ場合は 100% 広がるようにします 
  */
  width: 100%;
  height: 100%;
  min-height: 500px; /* ここで「表示されなくなる」のを防ぎます */
  background-color: #000;
  margin: 0;
  padding: 0;
  overflow: hidden;
  display: flex;
}

.vr-canvas-container {
  flex: 1;
  position: relative;
  width: 100%;
  height: 100%;
  cursor: grab;
}

.vr-canvas {
  /* canvasの下に隙間(4px程度)ができないように block 指定 */
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

.loader {
  border: 4px solid #333;
  border-top: 4px solid #3498db;
  border-radius: 50%;
  width: 30px;
  height: 30px;
  animation: spin 1s linear infinite;
}
@keyframes spin { 0% { transform: rotate(0deg); } 100% { transform: rotate(360deg); } }
</style>