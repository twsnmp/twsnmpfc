export default {
  search: {
    placeholder: 'キーワード...',
  },
  sort: {
    sortAsc: '昇順',
    sortDesc: '降順',
  },
  pagination: {
    previous: '前へ',
    next: '次へ',
    navigate: (page, pages) => `${page}/${pages}`,
    page: (page) => `${page}ページ`,
    showing: '表示中',
    of: 'まで',
    to: 'から',
    results: '件',
  },
  loading: '読み込み中...',
  noRecordsFound: '表示するデータがありません',
  error: 'データ読み込み中にエラーが発生しました',
};