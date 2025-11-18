import { writable,get } from 'svelte/store';

//セッション情報
export const session = writable({
	token: '',
  error: false,
});

// TWSNMPへログインする関数
export const login = async (user,password) => {
  try {
    const res = await fetch('APIURL/login', {
      method: 'POST',
      headers: {
      'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        UserID: user,
        Password: password,
      })
    })
    return await res.json();
  } catch (e) {
    console.log(e);
    return undefined;
  }
}

// twsnmpApiGetJSON : JSONをGETで取得するAPI
export const twsnmpApiGetJSON = async (api) => {
  const s = get(session);
  try {
    const res = await fetch('APIURL' + api, {
      method: 'GET',
      headers: {
      'Authorization': 'Bearer ' + s.token,
      },
    })
    if (res.status != 200) {
      return undefined;
    }
    return await res.json();
  } catch (e) {
    return undefined;
  }
}

// twsnmpApiPostJSON : POSTリクエスト
export const twsnmpApiPostJSON = async (api,data) => {
  const s = get(session);
  try {
    const res = await fetch('APIURL'+ api, {
      method: 'POST',
      headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer ' + s.token,
      },
      body: JSON.stringify(data),
    })
    return res.status == 204 || res.status == 200;
  } catch (e) {
    console.log(e);
    return false;
  }
}

export const twsnmpApiPostJSONWithData = async (api,data) => {
  const s = get(session);
  try {
    const res = await fetch('APIURL'+ api, {
      method: 'POST',
      headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer ' + s.token,
      },
      body: JSON.stringify(data),
    })
    if (res.status != 200) {
      return undefined;
    }
    return await res.json();
  } catch (e) {
    console.log(e);
    return false;
  }
}

export const twsnmpApiDelete = async (api) => {
  const s = get(session);
  try {
    const res = await fetch('APIURL'+ api, {
      method: 'DELETE',
      headers: {
      'Authorization': 'Bearer ' + s.token,
      },
    })
    return res.status == 204;
  } catch (e) {
    console.log(e);
    return false;
  }
}

// twsnmpApiUpload : ファイルのアップロード
export const twsnmpApiUpload = async (api,files) => {
  const s = get(session);
  const data = new FormData();
  data.append("file", files[0]);
  try {
    const res = await fetch('APIURL' + api, {
      method: 'POST',
      headers: {
      'Authorization': 'Bearer ' + s.token,
      },
      body: data,
    })
    return res.status == 204;
  } catch (e) {
    console.log(e);
    return false;
  }
}

// twsnmpApiDownload : ファイルのダウンロード
export const twsnmpApiDownload = async (api,file) => {
  const s = get(session);
  try {
    const res = await fetch('APIURL' + api, {
      method: 'GET',
      headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer ' + s.token,
      },
    })
    if (res.status != 200) {
      return false
    }
    const blob = await res.blob();
    var a = document.createElement("a");
    a.href = window.URL.createObjectURL(blob);
    a.download = file;
    a.click();
    a.remove();
  } catch (e) {
    console.log(e);
    return false;
  }
}
