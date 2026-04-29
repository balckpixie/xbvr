<template>
  <div class="vr-player-wrapper">
    <div class="vr-canvas-container">
      <canvas ref="vrCanvas" class="vr-canvas"></canvas>
      
      <div v-if="!isReady" class="vr-loading">
        <div class="loader"></div>
      </div>

      <div class="simple-controls">
        <button class="play-btn" @mousedown.stop @click="togglePlay">
          <span v-if="isPaused">▶ 再生</span>
          <span v-else>|| 一時停止</span>
        </button>
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
</style>