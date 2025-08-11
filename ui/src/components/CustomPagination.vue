<template>
  <div class="custom-pagination">
    <button
      :disabled="offset <= 0"
      @click="goPrevious"
      class="button is-small"
    >
      Previous
    </button>

    <span class="offset-display">{{ offset + 1 }} - {{ Math.min(offset + limit, total) }} / {{ total }}</span>

    <button
      :disabled="offset + limit >= total"
      @click="goNext"
      class="button is-small"
    >
      Next
    </button>
  </div>
</template>

<script>
export default {
  props: {
    total: { type: Number, required: true },    // 総件数
    limit: { type: Number, required: true },    // 1ページの最大表示件数
    offset: { type: Number, default: 0 },       // 現在のoffset（0始まり）
  },
  methods: {
    goPrevious() {
      if (this.offset > 0) {
        const newOffset = Math.max(0, this.offset - this.limit);
        this.$emit('change', newOffset);
      }
    },
    goNext() {
      if (this.offset + this.limit < this.total) {
        //const newOffset = Math.min(this.offset + this.limit, this.total - this.limit);
        const newOffset = this.offset + this.limit
        this.$emit('change', newOffset);
      }
    },
  },
};
</script>

<style scoped>
.custom-pagination {
  display: flex;
  align-items: center;
  gap: 0.75em;
}
.offset-display {
  min-width: 80px;
  text-align: center;
}
</style>
