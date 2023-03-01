import { getImageUrl } from '@misaon/imgproxy'

const KEY = '943b421c9eb07c830af81030552c86009268de4e532ba2ee2eab8247c6da0881'
const SALT = '520f986b998545b4785e0defbc4f3c1203f22de2374a3d53cb7a7fe9fea309c5'

const imageUrl = getImageUrl('https://img.example.com/pretty/image.jpg', {
    secret: KEY,
    salt: SALT,
    modifiers: {
        width: 300,
        height: 250,
    }
})

console.log(imageUrl)
