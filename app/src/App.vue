<script setup>
import Alert from './components/Alert.vue'
import { ref, onBeforeMount } from 'vue'
import { Icon } from '@iconify/vue'
import axios from 'axios'

const authorized = ref(false)

onBeforeMount(() => {
  if (typeof Telegram !== 'undefined' && Telegram.WebApp) {
    const tg = Telegram.WebApp;
    const user = tg.initDataUnsafe.user;

    if (user) {
      if (user.username == "incetro" || user.username == "corray9") authorized.value=true
      console.log("ID пользователя: ", user.id);
      console.log("Имя пользователя: ", user.first_name);
      console.log("Фамилия пользователя: ", user.last_name);
      console.log("Username: ", user.username);
      console.log("Язык: ", user.language_code);
    } else {
      console.log("Информация о пользователе недоступна.");
    }
  } else {
    console.log("Telegram Web App SDK не доступен.");
  }
})

const alerts = ref([])

const notionWaiting = ref(false)
const sheetsWaiting = ref(false)

const syncNotion = async () => {
  notionWaiting.value = true
  let msg = ""
  let status = "success"
  try {
    const test = await axios.patch(`${import.meta.env.VITE_API_URL}/sync`)
    msg = "Запущено обновление данных"
  } catch (err) {
    msg = "Данные уже обновляются"
    status = "error"
  }
  notionWaiting.value = false

  const id = Date.now()
  alerts.value.push({
    id,
    text: msg,
    color: status
  })
  setTimeout(() => {
    alerts.value = alerts.value.filter(alert => alert.id !== id)
  }, 3000)
}

const syncGSheets = async () => {
  sheetsWaiting.value = true
  let msg = ""
  let status = "success"
  try {
    await axios.patch(`${import.meta.env.VITE_API_URL}/sheets`)
    msg = "Данные обновлены"
  } catch (err) {
    console.log(err)
    msg = "Не удалось обновить данные"
    status = "error"
  }
  sheetsWaiting.value = false

  const id = Date.now()
  alerts.value.push({
    id,
    text: msg,
    color: status
  })
  setTimeout(() => {
    alerts.value = alerts.value.filter(alert => alert.id !== id)
  }, 3000)
}
</script>

<template>
  <div class="alerts absolute top-0 left-0 w-full p-5 flex justify-center">
    <transition-group name="bounce-fade">
      <Alert v-for="alert in alerts" :key="alert.id" :alert="alert" class="" />
    </transition-group>
  </div>
  <h1 v-if="!authorized">Куда это ты тянешь свои ручки?😑<br/>Я тебя не знаю, обратись к <a href="https://t.me/incetro" class=" underline">Андрею</a></h1>
  <section v-else class="p-5 flex flex-col items-center rounded-xl controls gap-2 bg-white w-full  max-w-96">
    <h1 class="font-bold text-2xl">Панель управления</h1>
    <section class="flex flex-col items-center w-full gap-2">
      <button @click="syncNotion" class="flex justify-center items-center">
        <Icon v-if="notionWaiting" class=" text-2xl" icon="eos-icons:three-dots-loading" />
        <p v-else>Синхронизировать Notion</p>
      </button>
      <button @click="syncGSheets" class="flex justify-center items-center">
        <Icon v-if="sheetsWaiting" class=" text-2xl" icon="eos-icons:three-dots-loading" />
        <p v-else>Синхронизировать Sheets</p>
      </button>
    </section>
  </section>
</template>

<style scoped>
.controls {
  box-shadow: 0.5rem 0.5rem 0rem 0rem rgb(0, 0, 0);
  border: solid 2px black;
}

@keyframes bounce-in {
  0% {
    transform: translateY(-100%);
    opacity: 0;
  }

  30% {
    transform: translateY(0);
    opacity: 1;
  }

  65% {
    transform: translateY(-10px);
  }

  100% {
    transform: translateY(0);
  }
}

@keyframes bounce-out {
  0% {
    transform: translateY(0);
    opacity: 1;
  }

  100% {
    transform: translateY(-100%);
    opacity: 0;
  }
}

.bounce-fade-enter-active {
  animation: bounce-in 0.5s ease-out;
}

.bounce-fade-leave-active {
  animation: bounce-out 0.5s ease-in;
}
</style>
