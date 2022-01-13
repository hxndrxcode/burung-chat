
const secure = location.protocol.indexOf('https') == 0
const wsURL = (secure ? 'wss': 'ws') + '://' + location.host + '/ws'
const vueApp = new Vue({
    el: '#root',
    data() {
        return {
            inputUsername: '',
            inputGroupName: '',
            username: '',
            currentRoom: {},
            inputText: '',
            roomList: [],
            ws: null,
            messageList: [],
            groupListRaw: [],
            userListRaw: [],
            roomListLoaded: false,
            showCreateGroupForm: false,
            retryWSConnect: 0,
            retryWSConnectLimit: 5,
            fetchingRoom: false,
        }
    },
    computed: {
        groupList() {
            if (!this.roomListLoaded) return []
            let joined = this.roomList.filter(v => v.Active == 1).map(v => v.ID)
            return this.groupListRaw.filter(v => !joined.includes(v.ID) )
        },
        userList() {
            if (!this.roomListLoaded) return []
            let chatted = this.roomList.filter(v => v.Type == 'p2p' && v.Active == 1)
            .map(v => v.Usernames)
            .map(v => {
                return v.find(w => w != this.username)
            })
            return this.userListRaw.filter(v => !chatted.includes(v.Username) && v.Username != this.username )
        }
    },
    mounted() {
        this.checkUser()
        this.bindEvent()
    },
    methods: {
        bindEvent() {
            const va = this
            window.addEventListener('popstate', event => {
                if (va.currentRoom.ID) {
                    va.closeRoom()
                }
            })
        },
        checkUser() {
            this.username = localStorage.getItem('username') || ''
            if (this.username) {
                this.fetchRoom()
                this.fetchHome()
                return
            }
        },
        login() {
            let username = this.inputUsername.toLowerCase().replace(/[^\w-]+/g, '')
            if (!username) {
                return alert('Username is required!')
            }
            const va = this
            $.ajax({
                url: '/login',
                method: 'post',
                data: {
                    username
                }
            })
            .then(res => {
                Vue.set(va, 'username', username)
                Vue.set(va, 'inputUsername', '')
                localStorage.setItem('username', username)
                va.fetchRoom()
                va.fetchHome()
            })
        },
        logout() {
            localStorage.removeItem('username')
            Vue.set(this, 'username', '')
            setTimeout(() => {
                this.checkUser()
            }, 500)
        },
        fetchHome() {
            const va = this
            $.ajax({
                url: '/group'
            }).then(res => {
                Vue.set(va, 'groupListRaw', res)
            })
            
            $.ajax({
                url: '/user'
            }).then(res => {
                Vue.set(va, 'userListRaw', res)
            })
        },
        fetchRoom(thenOpenRoomID) {
            this.fetchingRoom = true
            const va = this
            $.ajax({
                url: '/room',
                data: {
                    username: this.username
                }
            })
            .then(res => {
                Vue.set(va, 'roomList', res)
                if (thenOpenRoomID) {
                    va.openRoom(thenOpenRoomID)
                }
                Vue.set(va, 'roomListLoaded', true)
                Vue.set(va, 'fetchingRoom', false)
            })
            .catch(() => {
                Vue.set(va, 'roomListLoaded', true)
                Vue.set(va, 'fetchingRoom', false)
            })
        },
        joinRoom(room) {
            let find = this.roomList.find(v => v.ID == room && v.Active != 1)
            if (find) {
                this.openRoom(room)
                return
            }
            const va = this
            $.ajax({
                url: '/join',
                method: 'post',
                data: {
                    username: this.username,
                    room
                }
            }).then(res => {
                va.fetchRoom(room)
            })
        },
        openRoom(room) {
            if (this.currentRoom.ID == room) {
                return
            }
            if (this.currentRoom.ID != 'find') {
                this.closeRoom()
            }
            this.initWSConnection(room)
        },
        initWSConnection(room) {
            const va = this
            va.ws = new WebSocket(wsURL + '?room=' + room + '&username=' + this.username)
            va.ws.onopen = () => {
                va.retryWSConnect = 0
                if (va.currentRoom.ID != room) {
                    va.showRoom(room)
                }
            }
            va.ws.onmessage = event => {
                let msg = JSON.parse(event.data)
                this.messageList.push(msg)
                this.scrollToBottom()
            }
            va.ws.onclose = event => {
                if (va.retryWSConnect >= va.retryWSConnectLimit) {
                    alert('Internal server error. You are unable to receive and send a message')
                    return
                }
                va.retryWSConnect += 1
                va.initWSConnection(room)
            }
        },
        showRoom(roomID) {
            let room = this.roomList.find(v => v.ID == roomID)
            Vue.set(this, 'currentRoom', room)
            this.fetchMessage()
            history.pushState(null, '', location.href)
        },
        closeRoom() {
            if (this.ws) {
                this.ws.onclose = () => {}
                this.ws.close()
            }
            Vue.set(this, 'ws', null)
            Vue.set(this, 'currentRoom', {})
            Vue.set(this, 'messageList', [])
            this.fetchRoom()
            const va = this
            setTimeout(() => {
                $(va.$refs.findBtn).blur()
            }, 200);
        },
        findRoom() {
            this.currentRoom = {
                ID: 'find'
            }
            this.fetchHome()
            history.pushState(null, '', location.href)
        },
        fetchMessage() {
            let roomID = this.currentRoom.ID
            if (!roomID) {
                return
            }
            
            const va = this
            $.ajax({
                url: '/chat',
                data: {
                    room: roomID
                }
            }).then(res => {
                res = res.reverse()
                Vue.set(va, 'messageList', res)
                va.scrollToBottom()
            })
        },
        sendMessage() {
            let payload = {
                Message: this.inputText
            }
            this.ws.send(JSON.stringify(payload))
            this.inputText = ''
            this.scrollToBottom()
            if (this.currentRoom.Active != 1) {
                const va = this
                setTimeout(() => {
                    va.fetchRoom()
                }, 1000)
            }
        },
        scrollToBottom() {
            setTimeout(() => {
                const el = $('#msg-list')[0]
                el.scrollTop = el.scrollHeight
            }, 300)
        },
        toggleCreateGroup() {
            this.showCreateGroupForm = !this.showCreateGroupForm
            setTimeout(() => {
                this.$refs.inputGroupName.focus()
            }, 200)
        },
        createGroup() {
            let groupName = this.inputGroupName.replace(/ /g, '_').replace(/[^\w-]+/g, '')
            if (!groupName) {
                return
            }
            const va = this
            $.ajax({
                url: '/room',
                method: 'post',
                data: {
                    type: 'group',
                    user1: this.username,
                    label: groupName
                }
            }).then(res => {
                Vue.set(va, 'showCreateGroupForm', false)
                va.fetchRoom(res.ID)
                va.fetchHome()
            })
        },
        createP2PRoom(username) {
            let find = this.roomList.find(v => {
                return v.Usernames.includes(username) && v.Usernames.includes(this.username)
                    && v.Type == 'p2p' && v.Active != 1
            })
            if (find) {
                this.openRoom(find.ID)
                return
            }

            const va = this
            $.ajax({
                url: '/room',
                method: 'post',
                data: {
                    type: 'p2p',
                    user1: this.username,
                    user2: username,
                    label: this.username + '|' + username
                }
            }).then(res => {
                va.fetchRoom(res.ID)
                va.fetchHome()
            })
        },
        backBtn() {
            history.back()
        },
    },
    filters: {
        prettyDate: function (value) {
            if (!value) return ''
            const mArr = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
            const d = new Date(value)
            let date = d.getDate()
            let month = mArr[d.getMonth()]
            let hour = d.getHours()
            let min = d.getMinutes()
            return `${date} ${month}, ${hour}:${min}`
        },
        p2pLabel: function (value) {
            if (!value) return ''
            let myUsername = vueApp.username
            for (u of value.split('|')) {
                if (u != myUsername) {
                    return u
                }
            }
        }
    }
})