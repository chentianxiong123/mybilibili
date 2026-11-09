<template>
  <Page class="page">
    <ActionBar title="编辑资料" />
    <ScrollView>
      <StackLayout class="edit-content">
        <StackLayout class="form-group">
          <Label text="昵称" class="form-label" />
          <TextField v-model="nickname" class="form-input" />
        </StackLayout>
        <StackLayout class="form-group">
          <Label text="签名" class="form-label" />
          <TextView v-model="sign" class="form-textarea" />
        </StackLayout>
        <Button text="保存" class="save-btn" @tap="handleSave" />
      </StackLayout>
    </ScrollView>
  </Page>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'nativescript-vue'
import { $navigateBack } from 'nativescript-vue'
import { getUserInfo, updateUserInfo } from '../../api/user'
import storage from '../../utils/storage'

const nickname = ref('')
const sign = ref('')

onMounted(async () => {
  const user = storage.getUser<any>()
  if (user?.name) nickname.value = user.name
  try {
    const res = await getUserInfo()
    if (res.code === '1' && res.data) {
      nickname.value = res.data.name || nickname.value
      sign.value = res.data.sign || ''
    }
  } catch (e) {}
})

async function handleSave() {
  try {
    const res = await updateUserInfo({ name: nickname.value, sign: sign.value })
    if (res.code === '1') {
      $navigateBack()
    }
  } catch (e) {}
}
</script>

<style scoped>
.edit-content {
  padding: 20 16;
  background-color: white;
}

.form-group {
  margin-bottom: 16;
}

.form-label {
  font-size: 13;
  color: #61666d;
  margin-bottom: 6;
}

.form-input {
  width: 100%;
  height: 44;
  padding: 0 12;
  font-size: 15;
  border-width: 1;
  border-color: #e3e5e7;
  border-radius: 8;
  background-color: #f8f9fa;
}

.form-textarea {
  width: 100%;
  height: 80;
  padding: 8 12;
  font-size: 15;
  border-width: 1;
  border-color: #e3e5e7;
  border-radius: 8;
  background-color: #f8f9fa;
}

.save-btn {
  width: 100%;
  height: 44;
  background-color: #fb7299;
  color: white;
  font-size: 16;
  font-weight: bold;
  border-radius: 22;
  margin-top: 20;
}
</style>