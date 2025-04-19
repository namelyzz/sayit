export default function HelpPage() {
  return (
    <div className="max-w-3xl mx-auto">
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-8">
        <h1 className="text-3xl font-bold text-gray-900 mb-6">帮助与反馈</h1>

        <div className="prose prose-gray max-w-none space-y-6">
          <p className="text-gray-700 leading-relaxed">
            如果你在使用 SayIt 时遇到任何问题，或者有任何建议和反馈，欢迎通过以下方式联系我们：
          </p>

          <div className="bg-primary/5 border border-primary/20 rounded-lg p-6 mt-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">联系方式</h2>
            <div className="flex items-center space-x-3">
              <svg className="h-5 w-5 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
              </svg>
              <a href="mailto:905390065@qq.com" className="text-primary hover:underline text-lg">
                905390065@qq.com
              </a>
            </div>
          </div>

          <h2 className="text-xl font-semibold text-gray-900 mt-8">常见问题</h2>

          <div className="space-y-4">
            <div className="border-b border-gray-200 pb-4">
              <h3 className="font-medium text-gray-900 mb-2">如何注册账号？</h3>
              <p className="text-sm text-gray-600">
                点击页面右上角的"注册"按钮，填写用户名和密码即可完成注册。
              </p>
            </div>

            <div className="border-b border-gray-200 pb-4">
              <h3 className="font-medium text-gray-900 mb-2">如何发布帖子？</h3>
              <p className="text-sm text-gray-600">
                登录后，点击页面右上角的"发帖"按钮，选择社区、填写标题和内容后提交即可。
              </p>
            </div>

            <div className="border-b border-gray-200 pb-4">
              <h3 className="font-medium text-gray-900 mb-2">如何加入社区？</h3>
              <p className="text-sm text-gray-600">
                在左侧边栏浏览社区列表，点击任意社区即可查看该社区下的帖子。
              </p>
            </div>

            <div className="border-b border-gray-200 pb-4">
              <h3 className="font-medium text-gray-900 mb-2">投票规则是什么？</h3>
              <p className="text-sm text-gray-600">
                每个帖子支持赞成和反对投票，创建 7 天内的帖子可以投票。
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
