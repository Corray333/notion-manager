<script setup>

defineProps(["alert"])

</script>

<template>
    <section class=" w-full max-w-96 flex justify-center items-center p-5 rounded-xl bg-white border-dark absolute border-2" :class="`${alert.color}`">
        <p>{{ alert.text }}</p>
    </section>
</template>


<style></style>