<template>
  <div class="bg-light min-vh-100">
    <router-view />
  </div>
</template>

<style>
@import "https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/css/bootstrap.min.css";

body {
  font-family: "Inter", system-ui, -apple-system, sans-serif;
  background-color: #f8f9fa;
}

.card {
  border-radius: 1rem;
}

.shadow-sm {
  box-shadow: 0 0.125rem 0.25rem rgba(0, 0, 0, 0.075) !important;
}
</style>
