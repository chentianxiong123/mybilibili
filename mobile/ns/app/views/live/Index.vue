<template>
  <Page class="page">
    <ActionBar title="直播" />
    <ScrollView>
      <StackLayout>
        <StackLayout v-if="loading" class="loading">
          <Label text="加载中..." />
        </StackLayout>
        <StackLayout v-else>
          <StackLayout v-if="banners.length > 0" class="banner-section">
            <Image :src="banners[0]?.pic || ''" class="banner-img" />
          </StackLayout>

          <StackLayout v-if="areas.length > 0" class="areas-section">
            <Label text="直播分类" class="section-title" />
            <GridLayout columns="*, *, *, *" class="areas-grid">
              <StackLayout v-for="(area, i) in areas.slice(0, 8)" :key="i" col="i % 4" class="area-item" @tap="onAreaTap(area)">
                <Image :src="area.icon || ''" class="area-icon" />
                <Label :text="area.name || ''" class="area-name" />
              </StackLayout>
            </GridLayout>
          </StackLayout>

          <StackLayout v-if="liveRooms.length > 0" class="rooms-section">
            <Label text="正在直播" class="section-title" />
            <StackLayout v-for="(room, i) in liveRooms" :key="i" class="room-card" @tap="openRoom(room)">
              <Image :src="room.cover || room.user_cover || ''" class="room-cover" />
              <GridLayout columns="*, auto" class="room-meta">
                <Label :text="room.title || ''" class="room-title" col="0" textWrap="true" />
                <Label :text="'▶ ' + formatTenThousand(room.online || 0)" col="1" class="room-online" />
              </GridLayout>
              <GridLayout columns="auto, *" class="room-user">
                <Image :src="room.face || room.avatar || ''" class="room-avatar" col="0" />
                <Label :text="room.uname || room.nickname || ''" col="1" class="room-uname" />
              </GridLayout>
            </StackLayout>
          </StackLayout>
        </StackLayout>
      </StackLayout>
    </ScrollView>
  </Page>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'nativescript-vue'
import { $navigateTo, $showModal } from 'nativescript-vue'
import { getLiveIndexData, getLiveAreas } from '../../api/live'
import { formatTenThousand } from '../../utils/format'

const banners = ref<any[]>([])
const areas = ref<any[]>([])
const liveRooms = ref<any[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    const [indexRes, areasRes] = await Promise.all([
      getLiveIndexData(),
      getLiveAreas().catch(() => ({ code: '0', data: [] }))
    ])
    if (indexRes.code === '1') {
      const data = indexRes.data
      banners.value = data?.banners || data?.banner || []
      if (data?.itemList?.length > 0) {
        liveRooms.value = data.itemList[0]?.lives || []
      }
    }
    if (areasRes.code === '1') {
      areas.value = areasRes.data || []
    }
  } catch (e) {
    console.error('加载直播首页失败:', e)
  } finally {
    loading.value = false
  }
})

function onAreaTap(area: any) {
  $navigateTo(require('./Area.vue').default, {
    props: { areaId: area.id, areaName: area.name }
  })
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
.banner-section {
  padding: 6;
}

.banner-img {
  width: 100%;
  height: 160;
  border-radius: 5;
}

.areas-section {
  padding: 12 8;
  background-color: white;
  margin-top: 8;
}

.section-title {
  font-size: 15;
  font-weight: bold;
  color: #18191c;
  padding: 0 8 10;
}

.areas-grid {
  padding: 0 4;
}

.area-item {
  align-items: center;
  padding: 8 0;
}

.area-icon {
  width: 40;
  height: 40;
  border-radius: 20;
}

.area-name {
  font-size: 11;
  color: #61666d;
  margin-top: 4;
}

.rooms-section {
  padding: 12 8;
  background-color: white;
  margin-top: 8;
}

.room-card {
  padding: 8 4;
  border-bottom-width: 1;
  border-bottom-color: #f1f2f3;
}

.room-cover {
  width: 100%;
  height: 160;
  border-radius: 4;
}

.room-meta {
  margin-top: 6;
}

.room-title {
  font-size: 13;
  color: #18191c;
  font-weight: 500;
}

.room-online {
  font-size: 11;
  color: #fb7299;
}

.room-user {
  margin-top: 6;
  align-items: center;
}

.room-avatar {
  width: 24;
  height: 24;
  border-radius: 12;
  margin-right: 6;
}

.room-uname {
  font-size: 12;
  color: #9499a0;
}

.loading {
  padding: 40;
  text-align: center;
  color: #9499a0;
}
</style>