<template>
  <a class="button is-dark is-outlined is-small"
    @click="editActorImage(actor)"
    title="Edit Actor Image">
    <b-icon pack="mdi" icon="image-edit" size="is-small" />
  </a>
</template>

<script>
export default {
  name: 'ActorEditImageButton',
  props: { actor: Object },
  methods: {
    editActorImage (actor) {
      this.$store.commit('overlay/editActorImage', { actor })
    }
  }
}
</script>
