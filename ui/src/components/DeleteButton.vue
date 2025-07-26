<template>
  <a 
    :class="buttonClass"
    @click="deleteScene(item)"    
    :title="item.id == 0 ? 'Display scene details': 'Delete scene details'">
    <b-icon pack="mdi" icon="delete" size="is-small" />
  </a>
</template>

<script>
export default {
  name: 'DeleteButton',
  props: { item: Object },
  computed: {
    buttonClass () {
      if (this.item.is_deleted) {
        return 'button is-loading is-dark is-small'
      }
      return 'button is-danger is-outlined is-small'
    }
  },
  methods: {
    async deleteScene (scene) {
      let currentToggle=this.item.is_deleted
      const result = await this.$store.dispatch('sceneList/deleteScene', { scene });
      if (result) {
        this.item.is_deleted = !currentToggle;
      }
      this.$store.commit('overlay/hideDetails')
    }
  }
}
</script>
