<template>
  <Page class="page">
    <ActionBar title="直播列表">
      <NavigationButton text="◀" @tap="onBack" />
    </ActionBar>
    <ScrollView>
      <StackLayout>
        <StackLayout v-if="loading" class="loading">
          <Label text="加载中..." />
        </StackLayout>
        <GridLayout v-else-if="rooms.length > 0" columns="*, *" class="rooms-grid">
          <StackLayout v-for="(room, i) in rooms" :key="i" :col="i % 2" class="room-card" @tap="openRoom(room)">
            <Image :src="room.cover || room.user_cover || ''" class="room-cover" />
            <StackLayout class="room-info">
              <Label :text="room.title || ''" class="room-title" textWrap="true" />
              <GridLayout columns="auto, *" class="room-user-row">
                <Image :src="room.face || room.avatar || ''" class="room-avatar" col="0" />
                <Label :text="room.uname || room.nickname || ''" col="1" class="room-uname" />
              </GridLayout>
              <Label :text="'▶ ' + formatTenThousand(room.online || 0)" class="room-online" />
            </StackLayout>
          </StackLayout>
        </GridLayout>
        <StackLayout v-else class="empty-state">
          <Label text="暂无直播" />
        </StackLayout>
      </StackLayout>
    </ScrollView>
  </Page>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'nativescript-vue'
import { $navigateBack, $showModal } from 'nativescript-vue'
import { getLiveList } from '../../api/live'
import { formatTenThousand } from '../../utils/format'

const props = defineProps<{ category?: string }>()

const rooms = ref<any[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    const res = await getLiveList(props.category)
    if (res.code === '1') {
      rooms.value = res.data || []
    }
  } catch (e) {
    console.error('加载直播列表失败:', e)
  } finally {
    loading.value = false
  }
})

function onBack() {
  $navigateBack()
}

function openRoom(room: any) {
  const roomId = room.roomId || room.room_id || room.id
  if (roomId) {
    $showModal(require('./Room.vue').default, {
      fullscreen: true,
      props: { roomId }
    })
  }
}
</script>

<style scoped>
.rooms-grid {
  padding: 6;
}

.room-card {
  padding: 6;
}

.room-cover {
  width: 100%;
  height: 120;
  border-radius: 4;
}

.room-info {
  padding: 6 4;
  background-color: white;
}

.room-title {
  font-size: 13;
  color: #18191c;
  font-weight: 500;
  line-height: 1.3;
}

.room-user-row {
  margin-top: 4;
  align-items: center;
}

.room-avatar {
  width: 20;
  height: 20;
  border-radius: 10;
  margin-right: 4;
}

.room-uname {
  font-size: 11;
  color: #9499a0;
}

.room-online {
  font-size: 11;
  color: #fb7299;
  margin-top: 2;
}

.loading {
  padding: 40;
  text-align: center;
  color: #9499a0;
}

.empty-state {
  padding: 60 20;
  text-align: center;
  color: #9499a0;
  font-size: 14;
}
</style>