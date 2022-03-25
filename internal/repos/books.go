package repos

import "gitlab.com/egnd/bookshelf/internal/entities"

type BooksBleve struct {
	bleve entities.IBleveIndex
}

func NewBooksBleve(bleve entities.IBleveIndex) *BooksBleve {
	return &BooksBleve{
		bleve: bleve,
	}
}

// func (r *RecipesRepo) GetByID(ctx context.Context, id string) (*entities.Recipe, error) {
// 	res := &entities.Recipe{}
// 	err := r.db.WithContext(ctx).Preload("Tags").Take(res, id).Error

// 	return res, err
// }

// func (r *RecipesRepo) GetCount(ctx context.Context, tag *entities.Tag) (cnt int64, err error) {
// 	db := r.db.WithContext(ctx).
// 		Table("recipes r")

// 	if tag != nil {
// 		db = db.Joins("INNER JOIN recipes2tags rt ON r.id = rt.recipe_id AND rt.tag_id = ?", tag.ID)
// 	}

// 	err = db.Count(&cnt).Error

// 	return
// }

// func (r *RecipesRepo) GetList(ctx context.Context, tag *entities.Tag, sort entities.SortType, pager *pagination.Paginator) (res []entities.Recipe, err error) {
// 	var totalRecipes int64
// 	if totalRecipes, err = r.GetCount(ctx, tag); err != nil || totalRecipes == 0 {
// 		return
// 	}

// 	pager.SetNums(totalRecipes)

// 	var sortCond string
// 	switch sort {
// 	case entities.SortNewest:
// 		sortCond = "updated_at desc"
// 	case entities.SortPopular:
// 		sortCond = "visits desc"
// 	default:
// 		err = errors.New("invalid sort type")

// 		return
// 	}

// 	db := r.db.WithContext(ctx).Order(sortCond).Offset(pager.Offset()).Limit(pager.PerPageNums).Preload("Tags")

// 	if tag != nil {
// 		db = db.Joins("INNER JOIN recipes2tags rt ON recipes.id = rt.recipe_id AND rt.tag_id = ?", tag.ID)
// 	}

// 	err = db.Find(&res).Error

// 	return
// }

// func (r *RecipesRepo) GetSearchList(ctx context.Context, strQuery string, pager *pagination.Paginator) (res []entities.Recipe, err error) {
// 	query := bleve.NewMatchQuery(strQuery)
// 	cnt, _ := r.index.DocCount()
// 	search := bleve.NewSearchRequestOptions(query, int(cnt), 0, false)

// 	var searchResults *bleve.SearchResult
// 	if searchResults, err = r.index.Search(search); err != nil {
// 		return
// 	}

// 	if searchResults.Total == 0 {
// 		return
// 	}

// 	pager.SetNums(searchResults.Total)
// 	codes := make([]string, 0, pager.PerPageNums)

// 	for i := pager.Offset(); i < pager.Offset()+pager.PerPageNums; i++ {
// 		if len(searchResults.Hits) <= i {
// 			break
// 		}

// 		codes = append(codes, searchResults.Hits[i].ID)
// 	}

// 	tmpData := make([]entities.Recipe, 0, pager.PerPageNums)
// 	if err = r.db.WithContext(ctx).Preload("Tags").Find(&tmpData, "code in ?", codes).Error; err != nil {
// 		return
// 	}

// 	tmpIndex := map[string]entities.Recipe{}
// 	for _, item := range tmpData {
// 		tmpIndex[item.Code] = item
// 	}
// 	if len(tmpIndex) == 0 {
// 		return
// 	}

// 	for _, code := range codes {
// 		if item, ok := tmpIndex[code]; ok {
// 			res = append(res, item)
// 		}
// 	}

// 	return
// }

// func (r *RecipesRepo) GetSimilar(ctx context.Context, recipe entities.Recipe, limit int) (res []entities.Recipe, err error) {
// 	if len(recipe.Tags) == 0 {
// 		return
// 	}

// 	tagsID := make([]int, 0, len(recipe.Tags))
// 	for _, tag := range recipe.Tags {
// 		tagsID = append(tagsID, int(tag.ID))
// 	}

// 	err = r.db.WithContext(ctx).Order("updated_at desc").Limit(limit).
// 		Preload("Tags").
// 		Joins("INNER JOIN recipes2tags rt ON recipes.id = rt.recipe_id AND rt.tag_id IN ?", tagsID).
// 		Find(&res).Error

// 	return
// }
