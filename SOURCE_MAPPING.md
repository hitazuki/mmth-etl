# 来源映射关系

## Source ID 分配规则

| 范围 | 说明 |
|------|------|
| 0 | 未知/未匹配的来源 |
| 1-99999 | 游戏 TextResource ID |
| 100000+ | helper 自定义 ID |

---

## 可识别来源列表

### 游戏内置（部分匹配）

| Source ID | 别名 | 英文 | 繁中 | 简中 | 日文 | 韩文 |
|-----------|------|------|------|------|------|------|
| 140 | Fountain of Prayers | Fountain of Prayers: | 祈願之泉: | 祈愿之泉: | 祈りの泉: | 기원의 샘: |
| 67 | Open | Open | 開啟 | 开启 | 開く | 열기 |
| 719 | Login Bonus | Login | 簽到獎勵: | 签到奖励: | ログイン | ログ인 |
| 138 | Tower of Infinity | Tower of Infinity: | 無窮之塔: | 无穷之塔: | 無窮の塔: | 무한의 탑: |
| 2766 | Temple of Illusions | You have triumphed. | 勝利 | 胜利 | 勝利しました | 승리했습니다. |
| 21308 | Presents Box | Presents Box Claim All | 禮物箱 | 礼物箱 | プレゼントボックス | 선물 상자 |
| 21332 | Monthly Boost | Monthly Boost Already Claimed | 每月強化組合包 | 每月强化组合包 | 月間ブースト | 월간 부스트 |
| 3331 | Total Logins This Month | Total Logins This Month: | 本月累計簽到天數： | 本月累计签到天数： | 今月の合計ログイン日数： | 이번 달 보상 수령: |
| 23277 | World Player Clears | A player in your World | 本世界首次有玩家 | 本世界首次有玩家 | ワールド内のプレイヤーが初めて | 월드 내 플레이어가 최초로 |

### helper自定义（RewardMissionMsg 模式）

| Source ID | 别名 | 英文 | 繁中 | 简中 | 日文 | 韩文 |
|-----------|------|------|------|------|------|------|
| 23214 | Daily Mission Reward | Get Daily | 领取 Daily | 领取 Daily | Daily の | 일일 의 |
| 23215 | Weekly Mission Reward | Get Weekly | 领取 Weekly | 领取 Weekly | Weekly の | 주간 의 |
| 23213 | Main Mission Reward | Get Main | 领取 Main | 领取 Main | Main の | 메인 의 |

### helper自定义（自建 ID）

| Source ID | 别名 | 英文 | 中文 | 日文 | 韩文 |
|-----------|------|------|------|------|------|
| 100002 | Auto Buy Store Items | Auto Buy Store Items | 自动购买商城物品 | 自動購入ストアアイテム | 자동으로 상점 아이템 구매 |
| 100004 | Missions Claim All | Missions Claim All | 见下方多文本映射 | 见下方多文本映射 | 见下方多文本映射 |
| 100005 | Gacha | Gacha | 抽卡 | ガチャ | 가챠 |

**Missions Claim All (100004) 多文本映射**：

| 来源 | 英文 | 繁中 | 简中 | 日文 | 韩文 |
| ------ | ------ | ------ | ------ | ------ | ------ |
| 游戏 TextResource | You have no more challenges left. | 剩餘挑戰次數不足 | 剩余挑战次数不足 | 残り挑戦回数がありません | - |
| 游戏 TextResource | Cave of Space-TimeFinished | 時空洞窟已完成 | 时空洞窟已完成 | 時空の洞窟完了 | 시공의 동굴 완료 |
| helper ResourceStrings | Nothing to receive | 没有可以领取的 | | 受け取れるものはありません | 수령 가능한 것이 없습니다 |
| helper ResourceStrings | The expected diamond value of the current task is now below 20 | 当前任务的钻石数量期望值已低于20 | | 現在のタスクのダイヤの期待値が20未満になったため | 현재 작업의 다이아몬드 예상 값이 20 미만이므로 |

---

## 特殊来源提取规则

### Gacha 来源

Gacha 日志格式：`<prefix> <gacha_name> <count> <suffix>, ...`

**多语言前缀**：
| 语言 | 前缀 |
|------|------|
| EN | Gacha |
| TW/CN | 抽卡 |
| JA | ガチャ |
| KO | 가챠 |

**提取规则**：截取第二个空格之前的内容
- `Gacha 黒葬武具ガチャ 5 times` → `Gacha 黒葬武具ガチャ`
- `抽卡 黒葬武具ガチャ 5 次` → `抽卡 黒葬武具ガチャ`

### Open 来源

Open 日志格式：`<prefix> <item_name> x <count>`

**多语言前缀**：
| 语言 | 前缀 |
|------|------|
| EN | Open |
| TW | 開啟 |
| CN | 开启 |
| JA | 開く |
| KO | 열기 |

**提取规则**：截取 ` x` 之前的内容
- `Open Gold Sealed Chest x 5` → `Open Gold Sealed Chest`
- `開啟 上級封印寶箱 x 5` → `開啟 上級封印寶箱`

---

## 代码实现

### 来源映射表 (i18n/sources.go)

```go
// Game built-in source IDs from TextResource
const (
    SourceIDFountainOfPrayers SourceID = 140   // Fountain of Prayers -> 祈願之泉
    SourceIDOpen              SourceID = 67    // Open -> 開啟
    SourceIDLoginBonus        SourceID = 719   // Login Bonus -> 簽到獎勵
    SourceIDTowerInfinity     SourceID = 138   // Tower of Infinity -> 無窮之塔
    SourceIDTempleIllusions   SourceID = 2766  // Temple of Illusions -> 勝利
    SourceIDPresentsBox       SourceID = 21308 // Presents Box -> 禮物箱
    SourceIDMonthlyBoost      SourceID = 21332 // Monthly Boost -> 每月強化組合包
    SourceIDTotalLogins       SourceID = 3331  // Total Logins This Month -> 本月累計簽到天數
    SourceIDWorldClears       SourceID = 23277 // World Player Clears -> 本世界首次有玩家
)

// Helper custom source IDs (starting from 100002)
const (
    SourceIDAutoBuyStore    SourceID = 100002 // Auto Buy Store Items
    SourceIDMissionsClaimed SourceID = 100004 // Missions Claim All
    SourceIDGacha           SourceID = 100005 // Gacha (抽卡)
)

// MissionGroupType IDs from TextResource
const (
    MissionGroupDailyID  SourceID = 23214 // Daily -> 每日
    MissionGroupWeeklyID SourceID = 23215 // Weekly -> 每週
    MissionGroupMainID   SourceID = 23213 // Main -> 主線
)
```

### ChangeRecord 结构

```go
type ChangeRecord struct {
    Character string `json:"character"`
    Timestamp string `json:"timestamp"`
    Amount    int    `json:"amount"`
    Source    string `json:"source,omitempty"`
    SourceID  int    `json:"source_id,omitempty"` // 来源ID
}
```

### 输出示例

```json
{
  "character": "test",
  "timestamp": "2026-04-21T04:55:59+08:00",
  "amount": 100,
  "source": "Daily Mission Reward",
  "source_id": 23214
}
```

---

## 多语言支持

来源检测已支持以下语言：
- 英文 (EN)
- 繁中 (TW)
- 简中 (CN)
- 日文 (JA)
- 韩文 (KO)

---
生成时间: 2026-04-24
