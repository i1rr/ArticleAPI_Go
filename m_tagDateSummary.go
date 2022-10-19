package main

import "database/sql"

// TagDate - summary structure for tag and date, Articles limited by 10
type TagDate struct {
	Tag         string   `json:"tag"`
	Count       int      `json:"count"`
	Articles    []int    `json:"articles"`
	RelatedTags []string `json:"related_tags"`
}

func (td *TagDate) getArticleSummaryByTagAndDate(db *sql.DB, date string) (TagDate, error) {
	tagDate := TagDate{}
	tagDate.Tag = td.Tag //set tag given in url
	count := 0

	rows, err := db.Query(
		"SELECT id FROM articles a JOIN tags t ON a.id = t.article_id WHERE a.date = $1 AND t.tag = $2",
		date, td.Tag)

	if err != nil {
		return tagDate, err
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)

	tagDate.RelatedTags = append(tagDate.RelatedTags, td.Tag)
	for rows.Next() {
		if count < 10 { //limit to 10 articles as in requirements
			var id int

			if err := rows.Scan(&id); err != nil {
				return tagDate, err
			}

			tagDate.Articles = append(tagDate.Articles, id) //add articles according to requirements
		}
		count++ //keep count of articles
	}
	tagDate.RelatedTags = tagDate.RelatedTags[1:] //eliminating original tag
	tagDate.Count = count                         //set count of articles

	for _, id := range tagDate.Articles {
		rows, err := db.Query("SELECT tag FROM tags WHERE article_id = $1", id)

		if err != nil {
			return tagDate, err
		}

		for rows.Next() {
			var tag string
			if err := rows.Scan(&tag); err != nil {
				return tagDate, err
			}

			newTag := true

			for _, t := range tagDate.RelatedTags {
				if t == tag {
					newTag = false
					break
				}
			}

			if newTag {
				tagDate.RelatedTags = append(tagDate.RelatedTags, tag) //if tag is not in related tags, add it
			}
		}

		err = rows.Close()
		if err != nil {
			return TagDate{}, err
		}
	}
	return tagDate, err
}
