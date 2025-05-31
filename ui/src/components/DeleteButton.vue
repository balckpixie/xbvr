<template>
  <a class="button is-dark is-outlined is-small"
    @click="deleteScene(item)"    
    :title="item.id == 0 ? 'Delete scene details': 'Delete scene details'">
    <b-icon pack="mdi" icon="delete" size="is-small" />
  </a>
</template>

<script>
export default {
  name: 'DeleteButton',
  props: { item: Object },
  methods: {
    deleteScene (scene) {
      this.$store.commit('sceneList/deletescene', { scene })
    }
  }
}
</script>
